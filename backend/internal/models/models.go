package models

type SensorPayload struct {
	DeviceID string `json:"device_id"`
	Event    string `json:"event"`
	Details  string `json:"details"`

	Status     string  `json:"status"`
	DistanceCm float32 `json:"distance_cm"`
	LightLevel int     `json:"light_level"`

	Fails int    `json:"fails"`
	User  string `json:"user,omitempty"`

	RSSI   int     `json:"rssi"`
	Uptime float32 `json:"uptime"`

	RfidUID string `json:"rfid_uid,omitempty"`
}
