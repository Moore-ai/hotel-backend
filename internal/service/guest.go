package service

import (
	"errors"
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type GuestService struct {
	repo     *repository.GuestRepo
	auditLog *AuditLogService
}

func NewGuestService(repo *repository.GuestRepo, auditLog *AuditLogService) *GuestService {
	return &GuestService{repo: repo, auditLog: auditLog}
}

func (s *GuestService) FindByID(id uint) (*model.Guest, error) {
	return s.repo.FindByID(id)
}

func (s *GuestService) FindAll(page, pageSize int) ([]model.Guest, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *GuestService) Update(id uint, name, phone, email string) (*model.Guest, error) {
	guest, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldGuest := *guest
	if name != "" {
		guest.Name = name
	}
	if phone != "" {
		guest.Phone = phone
	}
	if email != "" {
		guest.Email = email
	}
	if err := s.repo.Update(guest); err != nil {
		return nil, err
	}
	if err := s.auditLog.Log(0, "updated", "guest", guest.ID, &oldGuest, guest, "修改客户"); err != nil {
		log.Printf("Audit log failed for guest %d: %v", guest.ID, err)
	}
	return guest, nil
}

var ErrGuestNotFound = errors.New("guest not found")
