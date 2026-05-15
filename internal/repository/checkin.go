package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type CheckinRepo struct {
	db *gorm.DB
}

func NewCheckinRepo(db *gorm.DB) *CheckinRepo {
	return &CheckinRepo{db: db}
}

func (r *CheckinRepo) Create(checkin *model.Checkin) error {
	return r.db.Create(checkin).Error
}

func (r *CheckinRepo) FindByID(id uint) (*model.Checkin, error) {
	var checkin model.Checkin
	err := r.db.Preload("User").Preload("Room").Preload("Order").First(&checkin, id).Error
	if err != nil {
		return nil, err
	}
	return &checkin, nil
}

func (r *CheckinRepo) FindAll(page, pageSize int) ([]model.Checkin, int64, error) {
	var checkins []model.Checkin
	var total int64
	r.db.Model(&model.Checkin{}).Count(&total)
	err := r.db.Preload("User").Preload("Room").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&checkins).Error
	return checkins, total, err
}

func (r *CheckinRepo) FindActive() ([]model.Checkin, error) {
	var checkins []model.Checkin
	err := r.db.Preload("User").Preload("Room").
		Where("status = ?", model.CheckinStatusActive).Find(&checkins).Error
	return checkins, err
}

func (r *CheckinRepo) Update(checkin *model.Checkin) error {
	return r.db.Save(checkin).Error
}

func (r *CheckinRepo) Delete(id uint) error {
	return r.db.Delete(&model.Checkin{}, id).Error
}
