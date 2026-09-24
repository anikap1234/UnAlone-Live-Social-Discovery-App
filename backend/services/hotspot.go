package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"unalone/backend/redis"
	"unalone/backend/utils"
	"unalone/backend/ws"
)

// Hotspots retain only transient aggregate centers/counts, never user positions.
type HotspotService struct {
	mu     sync.Mutex
	active map[string]ws.Hotspot
	hub    *ws.Hub
}

func NewHotspotService(hub *ws.Hub) *HotspotService {
	return &HotspotService{active: make(map[string]ws.Hotspot), hub: hub}
}
func (s *HotspotService) emit(kind string, h ws.Hotspot) {
	bytes, _ := json.Marshal(ws.HotspotEvent{Type: kind, Hotspot: h})
	s.hub.Broadcast(bytes)
}
func (s *HotspotService) reconcile(ctx context.Context) error {
	for id, h := range s.active {
		users, err := redis.NearbyUsers(ctx, h.Lat, h.Lon, 50)
		if err != nil {
			return err
		}
		if len(users) < 3 {
			delete(s.active, id)
			s.emit("HOTSPOT_DISSOLVED", h)
		} else {
			h.ActiveUsers = len(users)
			s.active[id] = h
		}
	}
	return nil
}
func (s *HotspotService) Reconcile(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reconcile(ctx)
}
func (s *HotspotService) Update(ctx context.Context, userID string, lat, lon float64) error {
	if err := redis.AddLocation(ctx, userID, lat, lon); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.reconcile(ctx); err != nil {
		return err
	}
	users, err := redis.NearbyPositions(ctx, lat, lon, 50)
	if err != nil {
		return err
	}
	if len(users) < 3 {
		return nil
	}
	// Publish a group centroid instead of a triggering user's raw GPS.
	centerLat, centerLon := 0.0, 0.0
	for _, user := range users {
		centerLat += user.Lat
		centerLon += user.Lon
	}
	centerLat /= float64(len(users))
	centerLon /= float64(len(users))
	members, err := redis.NearbyUsers(ctx, centerLat, centerLon, 50)
	if err != nil {
		return err
	}
	if len(members) < 3 {
		return nil
	}
	for _, existing := range s.active {
		if utils.DistanceMeters(existing.Lat, existing.Lon, centerLat, centerLon) <= 50 {
			return nil
		}
	}
	id := geohash(centerLat, centerLon)
	if _, exists := s.active[id]; exists {
		return nil
	}
	h := ws.Hotspot{ID: id, Lat: centerLat, Lon: centerLon, ActiveUsers: len(members)}
	s.active[id] = h
	s.emit("HOTSPOT_FORMED", h)
	return nil
}
func (s *HotspotService) Snapshot(ctx context.Context) ([]ws.Hotspot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.reconcile(ctx); err != nil {
		return nil, err
	}
	result := make([]ws.Hotspot, 0, len(s.active))
	for _, h := range s.active {
		result = append(result, h)
	}
	return result, nil
}

// StartExpiryListener listens to Redis events; it does not schedule cleanup jobs.
func (s *HotspotService) StartExpiryListener(ctx context.Context) error {
	channel := fmt.Sprintf("__keyevent@%d__:expired", redis.Client.Options().DB)
	subscription := redis.Client.Subscribe(ctx, channel)
	if _, err := subscription.Receive(ctx); err != nil {
		subscription.Close()
		return err
	}
	go func() {
		defer subscription.Close()
		go func() { <-ctx.Done(); subscription.Close() }()
		for {
			message, err := subscription.ReceiveMessage(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				// A missing event stream cannot justify displaying old hotspots.
				s.mu.Lock()
				for id, h := range s.active {
					delete(s.active, id)
					s.emit("HOTSPOT_DISSOLVED", h)
				}
				s.mu.Unlock()
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Second):
				}
				continue
			}
			if strings.HasPrefix(message.Payload, "user:") && strings.HasSuffix(message.Payload, ":presence") {
				timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
				if err := s.Reconcile(timeout); err != nil {
					log.Print("Could not reevaluate expired presence")
				}
				cancel()
			}
		}
	}()
	return nil
}

// Precision seven is the PRD's stable hotspot identity.
func geohash(lat, lon float64) string {
	const alphabet = "0123456789bcdefghjkmnpqrstuvwxyz"
	lo := []float64{-180, -90}
	hi := []float64{180, 90}
	values := []float64{lon, lat}
	result := make([]byte, 0, 7)
	value, bits := 0, 0
	for i := 0; i < 35; i++ {
		axis := i % 2
		mid := (lo[axis] + hi[axis]) / 2
		value <<= 1
		if values[axis] >= mid {
			value |= 1
			lo[axis] = mid
		} else {
			hi[axis] = mid
		}
		bits++
		if bits == 5 {
			result = append(result, alphabet[value])
			value = 0
			bits = 0
		}
	}
	return string(result)
}
