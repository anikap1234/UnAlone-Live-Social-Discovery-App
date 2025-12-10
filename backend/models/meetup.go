package models

import "time"

type Meetup struct {
    ID string `bson:"_id" json:"id"`
    Title string `bson:"title" json:"title"`
    Description string `bson:"description" json:"description"`
    Lat float64 `bson:"lat" json:"lat"`
    Lon float64 `bson:"lon" json:"lon"`
    CreatedBy string `bson:"createdBy" json:"createdBy"`
    CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
