package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"unalone/backend/db"
	"unalone/backend/models"
	"unalone/backend/utils"
)

func CreateMeetup(ctx context.Context, m *models.Meetup) error {
	m.ID = utils.New()
	m.CreatedAt = time.Now().Unix()
	_, err := db.Collection("meetups").InsertOne(ctx, m)
	return err
}

func NearbyMeetups(ctx context.Context, lat, lon, radius float64) ([]models.Meetup, error) {
	// Date filtering can use the time index; exact distance remains in Go.
	cursor, err := db.Collection("meetups").Find(ctx, bson.M{"time": bson.M{"$gte": time.Now().Unix()}}, options.Find().SetSort(bson.D{{Key: "time", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	result := make([]models.Meetup, 0)
	for cursor.Next(ctx) {
		var m models.Meetup
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		if utils.DistanceMeters(lat, lon, m.Lat, m.Lon) <= radius {
			result = append(result, m)
		}
	}
	return result, cursor.Err()
}
