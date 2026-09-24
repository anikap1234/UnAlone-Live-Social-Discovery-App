package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"unalone/backend/config"
)

var client *mongo.Client
var DB *mongo.Database

func Init(ctx context.Context, cfg *config.Config) error {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	opt := options.Client().ApplyURI(cfg.MongoURI)
	var err error
	client, err = mongo.Connect(ctx2, opt)
	if err != nil {
		return err
	}
	if err = client.Ping(ctx2, nil); err != nil {
		return err
	}
	DB = client.Database(cfg.MongoDB)
	_, err = DB.Collection("users").Indexes().CreateOne(ctx2, mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)})
	if err != nil {
		return err
	}
	_, err = DB.Collection("meetups").Indexes().CreateOne(ctx2, mongo.IndexModel{Keys: bson.D{{Key: "time", Value: 1}}})
	return err
}

func Disconnect(ctx context.Context) error {
	if client == nil {
		return nil
	}
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return client.Disconnect(ctx2)
}

func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}

func EnsureIndex(ctx context.Context, coll *mongo.Collection, keys interface{}) error {
	// DocumentDB-safe: create simple index
	_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: keys})
	return err
}

func URIPreview(cfg *config.Config) string {
	return fmt.Sprintf("mongodb: %s", cfg.MongoURI)
}
