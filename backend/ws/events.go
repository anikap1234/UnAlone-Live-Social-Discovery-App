package ws

type Hotspot struct {
	ID          string  `json:"hotspotId"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ActiveUsers int     `json:"activeUsers"`
}
type HotspotEvent struct {
	Type string `json:"type"`
	Hotspot
}
type MeetupCreated struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
