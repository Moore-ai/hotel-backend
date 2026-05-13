package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type CheckoutScheduleRepo struct {
	db *gorm.DB
}

func NewCheckoutScheduleRepo(db *gorm.DB) *CheckoutScheduleRepo {
	return &CheckoutScheduleRepo{db: db}
}

func (r *CheckoutScheduleRepo) Create(s *model.CheckoutSchedule) error {
	return r.db.Create(s).Error
}

func (r *CheckoutScheduleRepo) FindByCheckinID(checkinID uint) (*model.CheckoutSchedule, error) {
	var s model.CheckoutSchedule
	err := r.db.Where("checkin_id = ?", checkinID).First(&s).Error
	return &s, err
}

func (r *CheckoutScheduleRepo) Update(s *model.CheckoutSchedule) error {
	return r.db.Save(s).Error
}

func (r *CheckoutScheduleRepo) DeleteByCheckinID(checkinID uint) error {
	return r.db.Where("checkin_id = ?", checkinID).Delete(&model.CheckoutSchedule{}).Error
}
