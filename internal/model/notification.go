package model

import "time"

type Notification struct {
	ID               uint      `gorm:"primaryKey" json:"-"`
	NotificationCode string    `gorm:"-" json:"id"`
	UserID           uint      `gorm:"index;not null" json:"user_id"`
	User             *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Type             string    `gorm:"size:32;not null" json:"type"`
	Title            string    `gorm:"size:128;not null" json:"title"`
	Content          string    `gorm:"size:512" json:"content"`
	IsRead           bool      `gorm:"default:false" json:"is_read"`
	CreatedAt        time.Time `json:"created_at"`
}
