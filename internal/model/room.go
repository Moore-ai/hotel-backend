package model

import "time"

const (
	RoomStatusVacant   = "vacant"
	RoomStatusOccupied = "occupied"
	RoomStatusReserved = "reserved"
)

const (
	CheckinStatusActive     = "active"
	CheckinStatusCompleted  = "completed"
)

type Room struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RoomNumber    string    `gorm:"uniqueIndex;size:16;not null" json:"room_number"`
	Type          string    `gorm:"size:32" json:"type"`
	Capacity      int       `json:"capacity"`
	Floor         int       `json:"floor"`
	PricePerNight float64   `json:"price_per_night"`
	Status        string    `gorm:"size:16;default:vacant" json:"status"`
	Description   string    `gorm:"size:512" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
