package service

import (
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"gorm.io/gorm"
)

type OrderService struct {
	repo      *repository.OrderRepo
	roomRepo  *repository.RoomRepo
	allocator *RoomAllocator
	auditLog  *AuditLogService
	db        *gorm.DB
}

func NewOrderService(repo *repository.OrderRepo, roomRepo *repository.RoomRepo, allocator *RoomAllocator, auditLog *AuditLogService, db *gorm.DB) *OrderService {
	return &OrderService{repo: repo, roomRepo: roomRepo, allocator: allocator, auditLog: auditLog, db: db}
}

func (s *OrderService) Create(userID uint, roomID *uint, checkIn, checkOut string, price float64, guestCount int, roomType string) (*model.Order, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	roomRepo := s.roomRepo.WithTx(tx)
	orderRepo := s.repo.WithTx(tx)

	var allocatedRoom *model.Room
	var err error

	if roomID == nil {
		allocatedRoom, err = s.allocator.AllocateWithTx(tx, AllocateRequirements{
			GuestCount: guestCount,
			RoomType:   roomType,
		}, checkIn, checkOut)
		if err != nil {
			return nil, ErrNoRoomAvailable
		}
	} else {
		allocatedRoom, err = roomRepo.FindByID(*roomID)
		if err != nil {
			return nil, err
		}
	}

	order := &model.Order{
		UserID:       userID,
		RoomID:       allocatedRoom.ID,
		CheckInDate:  checkIn,
		CheckOutDate: checkOut,
		TotalPrice:   price,
		Status:       model.OrderStatusPending,
	}
	if err := orderRepo.Create(order); err != nil {
		return nil, err
	}

	allocatedRoom.Status = model.RoomStatusReserved
	if err := roomRepo.Update(allocatedRoom); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.auditLog.Log(0, "created", "order", order.ID, nil, order, "创建订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}
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
		switch order.Status {
		case model.OrderStatusConfirmed:
			action = "confirmed"
		case model.OrderStatusCancelled:
			action = "cancelled"
		}
	}
	if err := s.auditLog.Log(0, action, "order", order.ID, &oldOrder, order, "修改订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}
	return order, nil
}

func (s *OrderService) Delete(id uint) error {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if err := s.auditLog.Log(0, "deleted", "order", id, order, nil, "删除订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", id, err)
	}
	return s.repo.Delete(id)
}
