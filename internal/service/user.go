package service

import (
	"errors"

	"gorm.io/gorm"
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/hash"
)

type UserService struct {
	repo     *repository.UserRepo
	auditLog *AuditLogService
}

func NewUserService(repo *repository.UserRepo, auditLog *AuditLogService) *UserService {
	return &UserService{repo: repo, auditLog: auditLog}
}

func (s *UserService) Create(username, password, role, name, phone, email string) (*model.User, error) {
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	h, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: h,
		Role:         role,
		Name:         name,
		Phone:        phone,
		Email:        email,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "created", "user", user.ID, nil, user, "创建用户 "+user.Username)
	return user, nil
}

func (s *UserService) FindByID(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) FindAll(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *UserService) Update(id uint, username, password, role, name, phone, email string) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldUser := *user
	if username != "" {
		user.Username = username
	}
	if password != "" {
		h, err := hash.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = h
	}
	if role != "" {
		user.Role = role
	}
	if name != "" {
		user.Name = name
	}
	if phone != "" {
		user.Phone = phone
	}
	if email != "" {
		user.Email = email
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "updated", "user", user.ID, &oldUser, user, "修改用户 "+user.Username)
	return user, nil
}

func (s *UserService) Delete(id uint) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	s.auditLog.Log(0, "deleted", "user", id, user, nil, "删除用户")
	return s.repo.Delete(id)
}
