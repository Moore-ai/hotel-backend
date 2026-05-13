package service

import (
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type NotificationService struct {
	repo *repository.NotificationRepo
}

func NewNotificationService(repo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(userID uint, nType, title, content string) (*model.Notification, error) {
	n := &model.Notification{
		UserID:  userID,
		Type:    nType,
		Title:   title,
		Content: content,
	}
	return n, s.repo.Create(n)
}

func (s *NotificationService) FindByUserID(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	return s.repo.FindByUserID(userID, page, pageSize)
}

func (s *NotificationService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}

func (s *NotificationService) MarkRead(id, userID uint) error {
	return s.repo.MarkRead(id, userID)
}
