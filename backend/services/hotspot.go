package services

import (
    "context"
    "encoding/json"

    "unalone/backend/redis"
    "unalone/backend/ws"
)

func UpdateLocationAndMaybeBroadcast(ctx context.Context, hub *ws.Hub, email string, lat, lon float64) error {
    if err := redis.AddLocation(ctx, email, lat, lon); err != nil {
        return err
    }
    active, err := redis.NearbyUsers(ctx, lat, lon, 50)
    if err != nil {
        return err
    }
    if len(active) >= 3 {
        ev := ws.HotspotUpdate{Type: "HOTSPOT_UPDATE", Lat: lat, Lon: lon, ActiveUsers: len(active)}
        b, _ := json.Marshal(ev)
        hub.Broadcast(b)
    }
    return nil
}
