package repository

import "gorm.io/gorm"

type Repositories struct {
	User     *UserRepo
	Room     *RoomRepo
	Order    *OrderRepo
	Checkin  *CheckinRepo
	Notif    *NotificationRepo
	Audit    *AuditLogRepo
	Schedule *CheckoutScheduleRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:     NewUserRepo(db),
		Room:     NewRoomRepo(db),
		Order:    NewOrderRepo(db),
		Checkin:  NewCheckinRepo(db),
		Notif:    NewNotificationRepo(db),
		Audit:    NewAuditLogRepo(db),
		Schedule: NewCheckoutScheduleRepo(db),
	}
}
