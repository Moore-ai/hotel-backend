package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type GuestRepo struct {
	db *gorm.DB
}

func NewGuestRepo(db *gorm.DB) *GuestRepo {
	return &GuestRepo{db: db}
}

func (r *GuestRepo) Create(guest *model.Guest) error {
	return r.db.Create(guest).Error
}

func (r *GuestRepo) FindByID(id uint) (*model.Guest, error) {
	var guest model.Guest
	err := r.db.Preload("User").First(&guest, id).Error
	if err != nil {
		return nil, err
	}
	return &guest, nil
}

func (r *GuestRepo) FindByUserID(userID uint) (*model.Guest, error) {
	var guest model.Guest
	err := r.db.Where("user_id = ?", userID).First(&guest).Error
	if err != nil {
		return nil, err
	}
	return &guest, nil
}

func (r *GuestRepo) FindAll(page, pageSize int) ([]model.Guest, int64, error) {
	var guests []model.Guest
	var total int64
	r.db.Model(&model.Guest{}).Count(&total)
	err := r.db.Preload("User").Offset((page - 1) * pageSize).Limit(pageSize).Find(&guests).Error
	return guests, total, err
}

func (r *GuestRepo) Update(guest *model.Guest) error {
	return r.db.Save(guest).Error
}

func (r *GuestRepo) Delete(id uint) error {
	return r.db.Delete(&model.Guest{}, id).Error
}

func (r *GuestRepo) WithTx(tx *gorm.DB) *GuestRepo {
	return &GuestRepo{db: tx}
}
