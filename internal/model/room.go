package model

import "time"

type Room struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RoomNumber    string    `gorm:"uniqueIndex;size:16;not null" json:"room_number"`
	Type          string    `gorm:"size:32" json:"type"`
	Floor         int       `json:"floor"`
	PricePerNight float64   `json:"price_per_night"`
	Status        string    `gorm:"size:16;default:available" json:"status"`
	Description   string    `gorm:"size:512" json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
