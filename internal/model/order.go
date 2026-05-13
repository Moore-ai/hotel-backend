package model

import "time"

type Order struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RoomID       uint      `gorm:"index;not null" json:"room_id"`
	Room         *Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CheckInDate  string    `gorm:"size:16;not null" json:"check_in_date"`
	CheckOutDate string    `gorm:"size:16;not null" json:"check_out_date"`
	TotalPrice   float64   `json:"total_price"`
	Status       string    `gorm:"size:16;default:pending" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
