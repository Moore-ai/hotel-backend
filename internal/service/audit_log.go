package service

import (
	"encoding/json"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type AuditLogService struct {
	repo *repository.AuditLogRepo
}

func NewAuditLogService(repo *repository.AuditLogRepo) *AuditLogService {
	return &AuditLogService{repo: repo}
}

func (s *AuditLogService) Log(userID uint, action, entityType string, entityID uint, oldValue, newValue any, description string) error {
	var oldJSON, newJSON string
	if oldValue != nil {
		b, err := json.Marshal(oldValue)
		if err != nil {
			return err
		}
		oldJSON = string(b)
	}
	if newValue != nil {
		b, err := json.Marshal(newValue)
		if err != nil {
			return err
		}
		newJSON = string(b)
	}
	log := &model.AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		OldValue:    oldJSON,
		NewValue:    newJSON,
		Description: description,
	}
	return s.repo.Create(log)
}

func (s *AuditLogService) FindAll(f repository.AuditLogFilter) ([]model.AuditLog, int64, error) {
	return s.repo.FindAll(f)
}
