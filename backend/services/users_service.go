package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"unalone/backend/db"
	"unalone/backend/models"
	"unalone/backend/utils"
)

func FindOrCreateUser(ctx context.Context, email string) (*models.User, error) {
	collection := db.Collection("users")
	candidate := models.User{ID: utils.New(), Email: email, CreatedAt: time.Now().Unix()}
	_, err := collection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$setOnInsert": candidate}, options.Update().SetUpsert(true))
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		return nil, err
	}
	var user models.User
	if err = collection.FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
