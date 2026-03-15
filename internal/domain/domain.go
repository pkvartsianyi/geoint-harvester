package domain

import "time"

type GeoPoint struct {
	Type        string
	Coordinates []float64
}

type Message struct {
	Channel     string
	Content     string
	MsgID       int
	Timestamp   time.Time
	Geolocation *GeoPoint
}
