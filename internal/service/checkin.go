package service

import (
	"time"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type CheckinService struct {
	repo     *repository.CheckinRepo
	roomRepo *repository.RoomRepo
	auditLog *AuditLogService
}

func NewCheckinService(repo *repository.CheckinRepo, roomRepo *repository.RoomRepo, auditLog *AuditLogService) *CheckinService {
	return &CheckinService{repo: repo, roomRepo: roomRepo, auditLog: auditLog}
}

func (s *CheckinService) Create(orderID *uint, userID, roomID uint, expectedTime time.Time) (*model.Checkin, error) {
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, err
	}
	room.Status = "occupied"
	if err := s.roomRepo.Update(room); err != nil {
		return nil, err
	}

	checkin := &model.Checkin{
		OrderID:              orderID,
		UserID:               userID,
		RoomID:               roomID,
		CheckInTime:          time.Now(),
		ExpectedCheckoutTime: expectedTime,
		Status:               "active",
	}
	if err := s.repo.Create(checkin); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "checkin", "checkin", checkin.ID, nil, checkin, "办理入住")
	return checkin, nil
}

func (s *CheckinService) FindByID(id uint) (*model.Checkin, error) {
	return s.repo.FindByID(id)
}

func (s *CheckinService) FindAll(page, pageSize int) ([]model.Checkin, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *CheckinService) Checkout(id uint) (*model.Checkin, error) {
	checkin, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if checkin.Status != "active" {
		return nil, nil
	}

	now := time.Now()
	checkin.CheckOutTime = &now
	checkin.Status = "completed"

	if err := s.repo.Update(checkin); err != nil {
		return nil, err
	}

	room, _ := s.roomRepo.FindByID(checkin.RoomID)
	if room != nil {
		room.Status = "available"
		_ = s.roomRepo.Update(room)
	}

	s.auditLog.Log(0, "checkout", "checkin", checkin.ID, nil, checkin, "办理签离")
	return checkin, nil
}

func (s *CheckinService) Delete(id uint) error {
	return s.repo.Delete(id)
}
