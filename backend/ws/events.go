package ws

type HotspotUpdate struct {
    Type        string  `json:"type"`
    Lat         float64 `json:"lat"`
    Lon         float64 `json:"lon"`
    ActiveUsers int     `json:"activeUsers"`
}

type MeetupCreated struct {
    Type string      `json:"type"`
    Data interface{} `json:"data"`
}
