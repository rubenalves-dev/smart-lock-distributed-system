package mfa

import "time"

type MFARequest struct {
	ID             int       `json:"id"`
	RfidUID        string    `json:"rfid_uid"`
	DeviceID       string    `json:"device_id"`
	Fails          int       `json:"fails"`
	DistanceCm     float32   `json:"distance_cm"`
	LightLevel     int       `json:"light_level"`
	Classification int       `json:"classification"`
	Confidence     float32   `json:"confidence"`
	Recommendation string    `json:"recommendation"`
	Status         string    `json:"status"` // 'pending', 'approved', 'rejected'
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserName       *string   `json:"user_name,omitempty"`
}
