package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type RoomRepo struct {
	db *gorm.DB
}

func NewRoomRepo(db *gorm.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(room *model.Room) error {
	return r.db.Create(room).Error
}

func (r *RoomRepo) FindByID(id uint) (*model.Room, error) {
	var room model.Room
	err := r.db.First(&room, id).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepo) FindAll(page, pageSize int) ([]model.Room, int64, error) {
	var rooms []model.Room
	var total int64
	r.db.Model(&model.Room{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&rooms).Error
	return rooms, total, err
}

func (r *RoomRepo) Update(room *model.Room) error {
	return r.db.Save(room).Error
}

func (r *RoomRepo) FindByStatus(status string) ([]model.Room, error) {
	var rooms []model.Room
	err := r.db.Where("status = ?", status).Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepo) Delete(id uint) error {
	return r.db.Delete(&model.Room{}, id).Error
}

func (r *RoomRepo) FindAllRooms() ([]model.Room, error) {
	var rooms []model.Room
	err := r.db.Find(&rooms).Error
	return rooms, err
}

type RoomFilter struct {
	MinCapacity int
	Type        string
	Floor       int
}

func (r *RoomRepo) FindCandidates(filter RoomFilter) ([]model.Room, error) {
	var rooms []model.Room
	query := r.db.Model(&model.Room{})

	if filter.MinCapacity > 0 {
		query = query.Where("capacity >= ?", filter.MinCapacity)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Floor > 0 {
		query = query.Where("floor = ?", filter.Floor)
	}

	err := query.Find(&rooms).Error
	return rooms, err
}

func (r *RoomRepo) WithTx(tx *gorm.DB) *RoomRepo {
	return &RoomRepo{db: tx}
}
