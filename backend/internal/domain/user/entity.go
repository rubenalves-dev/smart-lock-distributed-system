package user

import "time"

type User struct {
	ID        int       `json:"id"`
	RfidUID   string    `json:"rfid_uid"`
	Name      *string   `json:"name"`
	Email     *string   `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
