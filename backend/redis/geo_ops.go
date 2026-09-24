package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rdb "github.com/redis/go-redis/v9"
	"time"
	"unalone/backend/utils"
)

const PresenceTTL = 30 * time.Second
const City = "local"

var ErrLocationRateLimited = errors.New("Location updates must be at least five seconds apart")

type Position struct {
	UserID string  `json:"userId"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
}

func PresenceKey(userID string) string { return "user:" + userID + ":presence" }

var locationScript = rdb.NewScript(`
if not redis.call('SET', KEYS[1], '1', 'NX', 'EX', 5) then return 0 end
local t=redis.call('TIME')
local bucket='geo:local:'..t[1]
redis.call('GEOADD', bucket, ARGV[1], ARGV[2], ARGV[3])
-- Fixed deadline: later writes must never prolong an older coordinate.
redis.call('EXPIREAT', bucket, tonumber(t[1])+30)
redis.call('SET', KEYS[2], ARGV[4], 'EX', 30)
return 1
`)

func AddLocation(ctx context.Context, userID string, lat, lon float64) error {
	if !utils.ValidCoordinates(lat, lon) {
		return errors.New("invalid coordinates")
	}
	value, err := json.Marshal(Position{UserID: userID, Lat: lat, Lon: lon})
	if err != nil {
		return err
	}
	ok, err := locationScript.Run(ctx, Client, []string{"location:rate:" + userID, PresenceKey(userID)}, lon, lat, userID, string(value)).Int()
	if err != nil {
		return err
	}
	if ok == 0 {
		return ErrLocationRateLimited
	}
	return nil
}

func RemoveLocation(ctx context.Context, userID string) error {
	// Coordinates in old buckets cannot contribute without current presence.
	return Client.Del(ctx, PresenceKey(userID)).Err()
}

func NearbyPositions(ctx context.Context, lat, lon, radius float64) ([]Position, error) {
	now, err := Client.Time(ctx).Result()
	if err != nil {
		return nil, err
	}
	pipeline := Client.Pipeline()
	queries := make([]*rdb.StringSliceCmd, 0, 31)
	for age := int64(0); age <= 30; age++ {
		key := fmt.Sprintf("geo:%s:%d", City, now.Unix()-age)
		queries = append(queries, pipeline.GeoSearch(ctx, key, &rdb.GeoSearchQuery{Latitude: lat, Longitude: lon, Radius: radius, RadiusUnit: "m"}))
	}
	if _, err := pipeline.Exec(ctx); err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, query := range queries {
		names, err := query.Result()
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			ids[name] = true
		}
	}
	keys := make([]string, 0, len(ids))
	for id := range ids {
		keys = append(keys, PresenceKey(id))
	}
	result := make([]Position, 0)
	if len(keys) == 0 {
		return result, nil
	}
	values, err := Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			continue
		}
		var position Position
		if err := json.Unmarshal([]byte(text), &position); err != nil {
			return nil, err
		}
		// Old bucket membership must never override a newer location.
		if utils.DistanceMeters(lat, lon, position.Lat, position.Lon) <= radius {
			result = append(result, position)
		}
	}
	return result, nil
}

func NearbyUsers(ctx context.Context, lat, lon, radius float64) ([]string, error) {
	positions, err := NearbyPositions(ctx, lat, lon, radius)
	users := make([]string, 0, len(positions))
	for _, position := range positions {
		users = append(users, position.UserID)
	}
	return users, err
}
