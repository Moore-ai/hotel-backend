package model

import "time"

type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name      string    `gorm:"size:64" json:"name"`
	Phone     string    `gorm:"size:32" json:"phone"`
	Email     string    `gorm:"size:128" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
