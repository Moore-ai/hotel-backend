package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type AuditLogRepo struct {
	db *gorm.DB
}

func NewAuditLogRepo(db *gorm.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

type AuditLogFilter struct {
	EntityType string
	EntityID   uint
	UserID     uint
	StartDate  string
	EndDate    string
	Page       int
	PageSize   int
}

func (r *AuditLogRepo) FindAll(f AuditLogFilter) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	q := r.db.Model(&model.AuditLog{}).Preload("User")
	if f.EntityType != "" {
		q = q.Where("entity_type = ?", f.EntityType)
	}
	if f.EntityID > 0 {
		q = q.Where("entity_id = ?", f.EntityID)
	}
	if f.UserID > 0 {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.StartDate != "" {
		q = q.Where("created_at >= ?", f.StartDate)
	}
	if f.EndDate != "" {
		q = q.Where("created_at <= ?", f.EndDate+" 23:59:59")
	}

	q.Count(&total)
	err := q.Offset((f.Page - 1) * f.PageSize).Limit(f.PageSize).
		Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}
