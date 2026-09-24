package models

type User struct {
	ID        string `bson:"_id" json:"userId"`
	Email     string `bson:"email" json:"email"`
	CreatedAt int64  `bson:"createdAt" json:"createdAt"`
}
