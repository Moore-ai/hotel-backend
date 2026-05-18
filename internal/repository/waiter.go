package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WaiterRepo struct {
	db *gorm.DB
}

func NewWaiterRepo(db *gorm.DB) *WaiterRepo {
	return &WaiterRepo{db: db}
}

func (r *WaiterRepo) Create(waiter *model.Waiter) error {
	return r.db.Create(waiter).Error
}

func (r *WaiterRepo) FindByID(id uint) (*model.Waiter, error) {
	var waiter model.Waiter
	err := r.db.Preload("User").Preload("ServingRoom").First(&waiter, id).Error
	if err != nil {
		return nil, err
	}
	return &waiter, nil
}

func (r *WaiterRepo) FindByUserID(userID uint) (*model.Waiter, error) {
	var waiter model.Waiter
	err := r.db.Where("user_id = ?", userID).First(&waiter).Error
	if err != nil {
		return nil, err
	}
	return &waiter, nil
}

func (r *WaiterRepo) FindAll(page, pageSize int) ([]model.Waiter, int64, error) {
	var waiters []model.Waiter
	var total int64
	r.db.Model(&model.Waiter{}).Count(&total)
	err := r.db.Preload("User").Preload("ServingRoom").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&waiters).Error
	return waiters, total, err
}

func (r *WaiterRepo) Update(waiter *model.Waiter) error {
	return r.db.Save(waiter).Error
}

func (r *WaiterRepo) Delete(id uint) error {
	return r.db.Delete(&model.Waiter{}, id).Error
}

func (r *WaiterRepo) WithTx(tx *gorm.DB) *WaiterRepo {
	return &WaiterRepo{db: tx}
}

func (r *WaiterRepo) ClearServingRoom(id uint) error {
	return r.db.Model(&model.Waiter{}).Where("id = ?", id).Update("serving_room_id", nil).Error
}

func (r *WaiterRepo) FindIdle() ([]model.Waiter, error) {
	var waiters []model.Waiter
	err := r.db.Preload("User").Where("serving_room_id IS NULL").Find(&waiters).Error
	return waiters, err
}

func (r *WaiterRepo) FindOneIdleWithLock(tx *gorm.DB) (*model.Waiter, error) {
	var waiter model.Waiter
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("User").Where("serving_room_id IS NULL").
		Order("RANDOM()").First(&waiter).Error
	return &waiter, err
}
