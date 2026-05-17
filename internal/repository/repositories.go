package repository

import "gorm.io/gorm"

type Repositories struct {
	db       *gorm.DB
	User     *UserRepo
	Guest    *GuestRepo
	Employee *EmployeeRepo
	Admin    *AdminRepo
	Room     *RoomRepo
	Order    *OrderRepo
	Checkin  *CheckinRepo
	Notif    *NotificationRepo
	Audit    *AuditLogRepo
	Schedule *CheckoutScheduleRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:       db,
		User:     NewUserRepo(db),
		Guest:    NewGuestRepo(db),
		Employee: NewEmployeeRepo(db),
		Admin:    NewAdminRepo(db),
		Room:     NewRoomRepo(db),
		Order:    NewOrderRepo(db),
		Checkin:  NewCheckinRepo(db),
		Notif:    NewNotificationRepo(db),
		Audit:    NewAuditLogRepo(db),
		Schedule: NewCheckoutScheduleRepo(db),
	}
}

func (r *Repositories) DB() *gorm.DB {
	return r.db
}
