package redis

import (
    "context"
    "fmt"
    "time"

    rdb "github.com/redis/go-redis/v9"
)

const LocationsKey = "locations"

// AddLocation adds or updates a user's geo position and sets a per-user TTL key
func AddLocation(ctx context.Context, user string, lat, lon float64) error {
    // GEOADD uses lon lat order
    if err := Client.GeoAdd(ctx, LocationsKey, &rdb.GeoLocation{Longitude: lon, Latitude: lat, Name: user}).Err(); err != nil {
        return err
    }
    // create per-user TTL key
    return Client.Set(ctx, fmt.Sprintf("loc_ttl:%s", user), "1", 30*time.Second).Err()
}

// NearbyUsers returns members within radiusMeters of lat/lon that have active TTL
func NearbyUsers(ctx context.Context, lat, lon float64, radiusMeters float64) ([]string, error) {
    // GeoSearchLocation is available via GeoSearch
    locs, err := Client.GeoSearch(ctx, LocationsKey, &rdb.GeoSearchQuery{
        Latitude:  lat,
        Longitude: lon,
        Radius:    radiusMeters,
        Unit:      "m",
        WithCoord: false,
        WithDist:  false,
        Count:     100,
        Sort:      "ASC",
    }).Result()
    if err != nil {
        return nil, err
    }
    var active []string
    for _, name := range locs {
        key := fmt.Sprintf("loc_ttl:%s", name)
        ok, _ := Client.Exists(ctx, key).Result()
        if ok > 0 {
            active = append(active, name)
        }
    }
    return active, nil
}
