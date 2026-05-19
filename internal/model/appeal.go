package model

import "time"

const (
	AppealStatusPending  = "pending"
	AppealStatusApproved = "approved"
	AppealStatusRejected = "rejected"
)

type Appeal struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"index;not null" json:"order_id"`
	Order      *Order    `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	User       *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Reason     string    `gorm:"type:text;not null" json:"reason"`
	Status     string    `gorm:"size:16;default:pending" json:"status"`
	ReviewerID *uint     `json:"reviewer_id"`
	Reviewer   *User     `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ReviewNote *string   `gorm:"type:text" json:"review_note"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
