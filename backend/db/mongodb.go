package db

import (
    "context"
    "fmt"
    "time"

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
    return nil
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
    _, err := coll.Indexes().CreateOne(ctx, keys)
    return err
}

func URIPreview(cfg *config.Config) string {
    return fmt.Sprintf("mongodb: %s", cfg.MongoURI)
}
