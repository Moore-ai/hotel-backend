package service

import (
	"encoding/json"
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type wsMessage struct {
	Type string       `json:"type"`
	Data *model.Notification `json:"data"`
}

const wsTypeNotification = "notification"

type NotificationService struct {
	repo *repository.NotificationRepo
	hub  *Hub
}

func NewNotificationService(repo *repository.NotificationRepo, hub *Hub) *NotificationService {
	return &NotificationService{repo: repo, hub: hub}
}

func (s *NotificationService) Create(userID uint, nType, title, content string) (*model.Notification, error) {
	n := &model.Notification{
		UserID:  userID,
		Type:    nType,
		Title:   title,
		Content: content,
	}
	if err := s.repo.Create(n); err != nil {
		return nil, err
	}

	msg, err := json.Marshal(wsMessage{Type: wsTypeNotification, Data: n})
	if err != nil {
		log.Printf("Failed to marshal notification for WebSocket: %v", err)
	} else {
		s.hub.SendToUser(userID, msg)
	}

	return n, nil
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
