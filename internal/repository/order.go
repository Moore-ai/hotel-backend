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

func (r *OrderRepo) FindOverlappingOrders(roomID uint, checkIn, checkOut string) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("room_id = ? AND status != ? AND check_in_date < ? AND check_out_date > ?",
		roomID, model.OrderStatusCancelled, checkOut, checkIn).
		Find(&orders).Error
	return orders, err
}

func (r *OrderRepo) FindOverlappingRoomIDs(roomIDs []uint, checkIn, checkOut string) ([]uint, error) {
	var result []uint
	err := r.db.Model(&model.Order{}).
		Distinct("room_id").
		Where("room_id IN ? AND status != ? AND check_in_date < ? AND check_out_date > ?",
			roomIDs, model.OrderStatusCancelled, checkOut, checkIn).
		Pluck("room_id", &result).Error
	return result, err
}

func (r *OrderRepo) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepo) FindByStatus(status string, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	r.db.Model(&model.Order{}).Where("status = ?", status).Count(&total)
	err := r.db.Preload("User").Preload("Room").
		Where("status = ?", status).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepo) Delete(id uint) error {
	return r.db.Delete(&model.Order{}, id).Error
}

func (r *OrderRepo) WithTx(tx *gorm.DB) *OrderRepo {
	return &OrderRepo{db: tx}
}
