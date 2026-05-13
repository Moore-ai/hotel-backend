package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepo) FindByUserID(userID uint, page, pageSize int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64
	r.db.Model(&model.Notification{}).Where("user_id = ?", userID).Count(&total)
	err := r.db.Where("user_id = ?", userID).
		Offset((page - 1) * pageSize).Limit(pageSize).
		Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *NotificationRepo) CountUnread(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *NotificationRepo) MarkRead(id, userID uint) error {
	return r.db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}
