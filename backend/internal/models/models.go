package models

type SensorPayload struct {
	DeviceID  string  `json:"device_id"`
	EventType string  `json:"event_type"`
	Value     float64 `json:"value"`
	IsLocked  bool    `json:"is_locked"`
}
