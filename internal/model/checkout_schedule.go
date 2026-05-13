package model

import "time"

type CheckoutSchedule struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	CheckinID       uint      `gorm:"uniqueIndex;not null" json:"checkin_id"`
	FirstThreshold  string    `gorm:"size:8;not null" json:"first_threshold"`
	SecondThreshold string    `gorm:"size:8;not null" json:"second_threshold"`
	FirstNotified   bool      `gorm:"default:false" json:"first_notified"`
	SecondNotified  bool      `gorm:"default:false" json:"second_notified"`
	CreatedAt       time.Time `json:"created_at"`
}
