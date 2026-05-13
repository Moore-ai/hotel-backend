package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepo) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("User").Preload("Room").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepo) FindAll(page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	r.db.Model(&model.Order{}).Count(&total)
	err := r.db.Preload("User").Preload("Room").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepo) FindByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	r.db.Model(&model.Order{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Preload("User").Preload("Room").
		Where("user_id = ?", userID).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepo) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepo) Delete(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}
