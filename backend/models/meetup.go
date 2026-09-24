package models

type Meetup struct {
	ID          string  `bson:"_id" json:"meetupId"`
	Title       string  `bson:"title" json:"title"`
	Description string  `bson:"description" json:"description"`
	Lat         float64 `bson:"lat" json:"lat"`
	Lon         float64 `bson:"lon" json:"lon"`
	Time        int64   `bson:"time" json:"time"`
	CreatedBy   string  `bson:"createdBy" json:"createdBy"`
	CreatedAt   int64   `bson:"createdAt" json:"createdAt"`
}
