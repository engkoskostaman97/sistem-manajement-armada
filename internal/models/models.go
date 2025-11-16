package models

type LocationMessage struct {
    VehicleID string  `json:"vehicle_id" db:"vehicle_id"`
    Latitude  float64 `json:"latitude" db:"latitude"`
    Longitude float64 `json:"longitude" db:"longitude"`
    Timestamp int64   `json:"timestamp" db:"ts"`
}

type GeofenceEvent struct {
    VehicleID string  `json:"vehicle_id"`
    Event     string  `json:"event"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Timestamp int64   `json:"timestamp"`
}
