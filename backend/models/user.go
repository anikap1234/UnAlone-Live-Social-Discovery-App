package models

import "time"

type User struct {
    ID        string    `bson:"_id" json:"id"`
    Email     string    `bson:"email" json:"email"`
    CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
