package services

import (
    "context"
    "math"
    "time"

    "go.mongodb.org/mongo-driver/bson"

    "unalone/backend/db"
    "unalone/backend/models"
    "unalone/backend/utils"
)

func CreateMeetup(ctx context.Context, m *models.Meetup) error {
    coll := db.Collection("meetups")
    m.ID = utils.New()
    m.CreatedAt = time.Now().UTC()
    _, err := coll.InsertOne(ctx, m)
    return err
}

func NearbyMeetups(ctx context.Context, lat, lon float64, maxMeters float64) ([]models.Meetup, error) {
    coll := db.Collection("meetups")
    // DocumentDB-safe: find all and filter in-app by distance (could be optimized)
    cur, err := coll.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var out []models.Meetup
    for cur.Next(ctx) {
        var m models.Meetup
        if err := cur.Decode(&m); err != nil {
            continue
        }
        if distanceMeters(lat, lon, m.Lat, m.Lon) <= maxMeters {
            out = append(out, m)
        }
    }
    return out, nil
}

func distanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371000.0
    toRad := func(d float64) float64 { return d * math.Pi / 180 }
    dLat := toRad(lat2 - lat1)
    dLon := toRad(lon2 - lon1)
    a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}
