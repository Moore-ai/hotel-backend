package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type AdminRepo struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepo {
	return &AdminRepo{db: db}
}

func (r *AdminRepo) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepo) FindByID(id uint) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.Preload("User").First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepo) FindByUserID(userID uint) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.Where("user_id = ?", userID).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepo) FindAll(page, pageSize int) ([]model.Admin, int64, error) {
	var admins []model.Admin
	var total int64
	r.db.Model(&model.Admin{}).Count(&total)
	err := r.db.Preload("User").Offset((page - 1) * pageSize).Limit(pageSize).Find(&admins).Error
	return admins, total, err
}

func (r *AdminRepo) Update(admin *model.Admin) error {
	return r.db.Save(admin).Error
}

func (r *AdminRepo) Delete(id uint) error {
	return r.db.Delete(&model.Admin{}, id).Error
}

func (r *AdminRepo) WithTx(tx *gorm.DB) *AdminRepo {
	return &AdminRepo{db: tx}
}
