package service

import (
	"encoding/json"
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/obfuscate"

	"github.com/speps/go-hashids/v2"
)

type wsMessage struct {
	Type string             `json:"type"`
	Data *model.Notification `json:"data"`
}

const wsTypeNotification = "notification"

type NotificationService struct {
	repo         *repository.NotificationRepo
	hub          *Hub
	obfuscateKey *hashids.HashID
}

func NewNotificationService(repo *repository.NotificationRepo, hub *Hub, obfuscateKey *hashids.HashID) *NotificationService {
	return &NotificationService{repo: repo, hub: hub, obfuscateKey: obfuscateKey}
}

func (s *NotificationService) DecodeCode(code string) (uint, error) {
	return obfuscate.Decode(s.obfuscateKey, code)
}

func (s *NotificationService) encodeNotification(n *model.Notification) {
	code, err := obfuscate.Encode(s.obfuscateKey, n.ID)
	if err != nil {
		log.Printf("Failed to encode notification %d: %v", n.ID, err)
		return
	}
	n.NotificationCode = code
}

func (s *NotificationService) encodeNotifications(ns []model.Notification) {
	for i := range ns {
		s.encodeNotification(&ns[i])
	}
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

	s.encodeNotification(n)

	msg, err := json.Marshal(wsMessage{Type: wsTypeNotification, Data: n})
	if err != nil {
		log.Printf("Failed to marshal notification for WebSocket: %v", err)
	} else {
		s.hub.SendToUser(userID, msg)
	}

	return n, nil
}

func (s *NotificationService) FindByUserID(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	ns, total, err := s.repo.FindByUserID(userID, page, pageSize)
	if err == nil {
		s.encodeNotifications(ns)
	}
	return ns, total, err
}

func (s *NotificationService) CountUnread(userID uint) (int64, error) {
	return s.repo.CountUnread(userID)
}

func (s *NotificationService) MarkRead(id, userID uint) error {
	return s.repo.MarkRead(id, userID)
}
