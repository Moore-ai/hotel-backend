package service

import (
	"errors"
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type AdminService struct {
	repo     *repository.AdminRepo
	auditLog *AuditLogService
}

func NewAdminService(repo *repository.AdminRepo, auditLog *AuditLogService) *AdminService {
	return &AdminService{repo: repo, auditLog: auditLog}
}

func (s *AdminService) FindByID(id uint) (*model.Admin, error) {
	return s.repo.FindByID(id)
}

func (s *AdminService) FindAll(page, pageSize int) ([]model.Admin, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *AdminService) Update(id uint, name, phone, email string) (*model.Admin, error) {
	admin, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldAdmin := *admin
	if name != "" {
		admin.Name = name
	}
	if phone != "" {
		admin.Phone = phone
	}
	if email != "" {
		admin.Email = email
	}
	if err := s.repo.Update(admin); err != nil {
		return nil, err
	}
	if err := s.auditLog.Log(0, "updated", "admin", admin.ID, &oldAdmin, admin, "修改管理员"); err != nil {
		log.Printf("Audit log failed for admin %d: %v", admin.ID, err)
	}
	return admin, nil
}

var ErrAdminNotFound = errors.New("admin not found")
