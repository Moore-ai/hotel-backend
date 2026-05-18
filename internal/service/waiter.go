package service

import (
	"errors"
	"log"
	"math/rand"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"gorm.io/gorm"
)

type WaiterService struct {
	repo     *repository.WaiterRepo
	notifSvc *NotificationService
	auditLog *AuditLogService
	db       *gorm.DB
}

func NewWaiterService(repo *repository.WaiterRepo, notifSvc *NotificationService, auditLog *AuditLogService, db *gorm.DB) *WaiterService {
	return &WaiterService{repo: repo, notifSvc: notifSvc, auditLog: auditLog, db: db}
}

func (s *WaiterService) FindByID(id uint) (*model.Waiter, error) {
	return s.repo.FindByID(id)
}

func (s *WaiterService) FindByUserID(userID uint) (*model.Waiter, error) {
	return s.repo.FindByUserID(userID)
}

func (s *WaiterService) FindAll(page, pageSize int) ([]model.Waiter, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *WaiterService) Update(id uint, name, phone, email string) (*model.Waiter, error) {
	waiter, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldWaiter := *waiter
	if name != "" {
		waiter.Name = name
	}
	if phone != "" {
		waiter.Phone = phone
	}
	if email != "" {
		waiter.Email = email
	}
	if err := s.repo.Update(waiter); err != nil {
		return nil, err
	}
	if err := s.auditLog.Log(0, "updated", "waiter", waiter.ID, &oldWaiter, waiter, "修改服务员"); err != nil {
		log.Printf("Audit log failed for waiter %d: %v", waiter.ID, err)
	}
	return waiter, nil
}

func (s *WaiterService) Dispatch(roomID uint, guestUserID uint, serviceType, note string) (*model.Waiter, error) {
	// 查找空闲服务员
	idle, err := s.repo.FindIdle()
	if err != nil {
		return nil, err
	}
	if len(idle) == 0 {
		return nil, ErrNoWaiterAvailable
	}

	// 随机派单
	chosen := &idle[rand.Intn(len(idle))]
	chosen.ServingRoomID = &roomID
	if err := s.repo.Update(chosen); err != nil {
		return nil, err
	}

	// 通知客户
	content := ""
	if chosen.Name != "" {
		content = chosen.Name
	} else if chosen.User != nil {
		content = chosen.User.Username
	} else {
		content = "服务员"
	}
	if chosen.Phone != "" {
		content += "（电话：" + chosen.Phone + "）"
	}
	content += "正在前往您的房间"
	if serviceType != "" {
		content += "处理「" + serviceType + "」"
	}
	if note != "" {
		content += "（备注：" + note + "）"
	}
	content += "，请稍候。"

	if _, err := s.notifSvc.Create(guestUserID, "waiter_assigned", "服务员已派单", content); err != nil {
		log.Printf("Notification failed for user %d: %v", guestUserID, err)
	}

	return chosen, nil
}

func (s *WaiterService) NotifyWaiting(guestUserID uint) {
	if _, err := s.notifSvc.Create(guestUserID, "waiter_unavailable", "服务员忙碌",
		"当前暂无空闲服务员，请稍候，我们会在有空闲服务员时为您安排。"); err != nil {
		log.Printf("Notification failed for user %d: %v", guestUserID, err)
	}
}

func (s *WaiterService) CompleteService(id uint) error {
	return s.repo.ClearServingRoom(id)
}

var ErrNoWaiterAvailable = errors.New("no waiter available")
