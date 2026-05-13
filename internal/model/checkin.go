package model

import "time"

type Checkin struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	OrderID              *uint     `gorm:"index" json:"order_id"`
	Order                *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	UserID               uint      `gorm:"index;not null" json:"user_id"`
	User                 *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RoomID               uint      `gorm:"index;not null" json:"room_id"`
	Room                 *Room     `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	CheckInTime          time.Time `json:"check_in_time"`
	CheckOutTime         *time.Time `json:"check_out_time"`
	ExpectedCheckoutTime time.Time `json:"expected_checkout_time"`
	Status               string    `gorm:"size:16;default:active" json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
