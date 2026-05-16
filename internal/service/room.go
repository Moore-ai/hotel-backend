package service

import (
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type RoomService struct {
	repo     *repository.RoomRepo
	auditLog *AuditLogService
}

func NewRoomService(repo *repository.RoomRepo, auditLog *AuditLogService) *RoomService {
	return &RoomService{repo: repo, auditLog: auditLog}
}

func (s *RoomService) Create(roomNumber, roomType string, capacity, floor int, price float64, status, desc string) (*model.Room, error) {
	if status == "" {
		status = model.RoomStatusVacant
	}
	room := &model.Room{
		RoomNumber:    roomNumber,
		Type:          roomType,
		Capacity:      capacity,
		Floor:         floor,
		PricePerNight: price,
		Status:        status,
		Description:   desc,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "created", "room", room.ID, nil, room, "创建房间 "+room.RoomNumber)
	return room, nil
}

func (s *RoomService) FindByID(id uint) (*model.Room, error) {
	return s.repo.FindByID(id)
}

func (s *RoomService) FindAll(page, pageSize int) ([]model.Room, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *RoomService) Update(id uint, roomNumber, roomType string, capacity, floor int, price float64, status, desc string) (*model.Room, error) {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldRoom := *room
	if roomNumber != "" {
		room.RoomNumber = roomNumber
	}
	if roomType != "" {
		room.Type = roomType
	}
	if capacity > 0 {
		room.Capacity = capacity
	}
	if floor > 0 {
		room.Floor = floor
	}
	if price > 0 {
		room.PricePerNight = price
	}
	if status != "" {
		room.Status = status
	}
	if desc != "" {
		room.Description = desc
	}
	if err := s.repo.Update(room); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "updated", "room", room.ID, &oldRoom, room, "修改房间 "+room.RoomNumber)
	return room, nil
}

func (s *RoomService) Delete(id uint) error {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	s.auditLog.Log(0, "deleted", "room", id, room, nil, "删除房间")
	return s.repo.Delete(id)
}
