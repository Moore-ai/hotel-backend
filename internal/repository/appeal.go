package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type AppealRepo struct {
	db *gorm.DB
}

func NewAppealRepo(db *gorm.DB) *AppealRepo {
	return &AppealRepo{db: db}
}

func (r *AppealRepo) Create(a *model.Appeal) error {
	return r.db.Create(a).Error
}

func (r *AppealRepo) FindByID(id uint) (*model.Appeal, error) {
	var a model.Appeal
	err := r.db.Preload("Order").Preload("User").First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AppealRepo) FindByOrderID(orderID uint) (*model.Appeal, error) {
	var a model.Appeal
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AppealRepo) FindByStatus(status string, page, pageSize int) ([]model.Appeal, int64, error) {
	var list []model.Appeal
	var total int64
	query := r.db.Model(&model.Appeal{}).Preload("Order").Preload("User").Preload("Reviewer")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *AppealRepo) FindByUserID(userID uint, status string, page, pageSize int) ([]model.Appeal, int64, error) {
	var list []model.Appeal
	var total int64
	query := r.db.Model(&model.Appeal{}).Where("user_id = ?", userID).Preload("Order").Preload("User").Preload("Reviewer")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *AppealRepo) Update(a *model.Appeal) error {
	return r.db.Save(a).Error
}

func (r *AppealRepo) FindByIDForUpdate(tx *gorm.DB, id uint) (*model.Appeal, error) {
	var a model.Appeal
	err := tx.Preload("Order").Preload("User").First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AppealRepo) WithTx(tx *gorm.DB) *AppealRepo {
	return &AppealRepo{db: tx}
}
