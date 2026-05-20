package model

import "time"

type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Name      string    `gorm:"size:64" json:"name"`
	Phone     string    `gorm:"size:32" json:"phone"`
	Email     string    `gorm:"size:128" json:"email"`
	HireDate  string    `gorm:"size:16" json:"hire_date"`
	Salary    float64   `gorm:"default:0" json:"salary"`
	Notes     string    `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
