package service

import (
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type OrderService struct {
	repo     *repository.OrderRepo
	auditLog *AuditLogService
}

func NewOrderService(repo *repository.OrderRepo, auditLog *AuditLogService) *OrderService {
	return &OrderService{repo: repo, auditLog: auditLog}
}

func (s *OrderService) Create(userID, roomID uint, checkIn, checkOut string, price float64) (*model.Order, error) {
	order := &model.Order{
		UserID:       userID,
		RoomID:       roomID,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		TotalPrice:   price,
		Status:       "pending",
	}
	if err := s.repo.Create(order); err != nil {
		return nil, err
	}
	s.auditLog.Log(0, "created", "order", order.ID, nil, order, "创建订单")
	return order, nil
}

func (s *OrderService) FindByID(id uint) (*model.Order, error) {
	return s.repo.FindByID(id)
}

func (s *OrderService) FindAll(page, pageSize int) ([]model.Order, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *OrderService) FindByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	return s.repo.FindByUserID(userID, page, pageSize)
}

func (s *OrderService) Update(id uint, checkIn, checkOut, status string, price float64) (*model.Order, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldOrder := *order
	if checkIn != "" {
		order.CheckInDate = checkIn
	}
	if checkOut != "" {
		order.CheckOutDate = checkOut
	}
	if status != "" {
		order.Status = status
	}
	if price > 0 {
		order.TotalPrice = price
	}
	if err := s.repo.Update(order); err != nil {
		return nil, err
	}
	action := "updated"
	if oldOrder.Status != order.Status {
		if order.Status == "confirmed" {
			action = "confirmed"
		} else if order.Status == "cancelled" {
			action = "cancelled"
		}
	}
	s.auditLog.Log(0, action, "order", order.ID, &oldOrder, order, "修改订单")
	return order, nil
}

func (s *OrderService) Delete(id uint) error {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	s.auditLog.Log(0, "deleted", "order", id, order, nil, "删除订单")
	return s.repo.Delete(id)
}
