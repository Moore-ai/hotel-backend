package service

import (
	"log"
	"time"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"gorm.io/gorm"
)

type CheckinService struct {
	repo      *repository.CheckinRepo
	roomRepo  *repository.RoomRepo
	orderRepo *repository.OrderRepo
	allocator *RoomAllocator
	auditLog  *AuditLogService
	db        *gorm.DB
}

func NewCheckinService(repo *repository.CheckinRepo, roomRepo *repository.RoomRepo, orderRepo *repository.OrderRepo, allocator *RoomAllocator, auditLog *AuditLogService, db *gorm.DB) *CheckinService {
	return &CheckinService{repo: repo, roomRepo: roomRepo, orderRepo: orderRepo, allocator: allocator, auditLog: auditLog, db: db}
}

func (s *CheckinService) Create(orderID *uint, userID, roomID uint, expectedTime time.Time) (*model.Checkin, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	roomRepo := s.roomRepo.WithTx(tx)
	orderRepo := s.orderRepo.WithTx(tx)
	checkinRepo := s.repo.WithTx(tx)

	var allocatedRoom *model.Room
	var err error
	var actualOrderID *uint

	if orderID == nil && roomID == 0 {
		allocatedRoom, err = s.allocator.AllocateWithTx(tx, AllocateRequirements{}, time.Now().Format("2006-01-02"), expectedTime.Format("2006-01-02"))
		if err != nil {
			return nil, ErrNoRoomAvailable
		}

		order := &model.Order{
			UserID:       userID,
			RoomID:       allocatedRoom.ID,
			CheckInDate:  time.Now().Format("2006-01-02"),
			CheckOutDate: expectedTime.Format("2006-01-02"),
			TotalPrice:   0,
			Status:       model.OrderStatusConfirmed,
		}
		if err := orderRepo.Create(order); err != nil {
			return nil, err
		}
		actualOrderID = &order.ID
	} else if orderID != nil {
		order, err := orderRepo.FindByID(*orderID)
		if err != nil {
			return nil, err
		}
		allocatedRoom, err = roomRepo.FindByID(order.RoomID)
		if err != nil {
			return nil, err
		}
		actualOrderID = orderID
	} else {
		allocatedRoom, err = roomRepo.FindByID(roomID)
		if err != nil {
			return nil, err
		}
		actualOrderID = nil
	}

	allocatedRoom.Status = model.RoomStatusOccupied
	if err := roomRepo.Update(allocatedRoom); err != nil {
		return nil, err
	}

	checkin := &model.Checkin{
		OrderID:              actualOrderID,
		UserID:               userID,
		RoomID:               allocatedRoom.ID,
		CheckInTime:          time.Now(),
		ExpectedCheckoutTime: expectedTime,
		Status:               model.CheckinStatusActive,
	}
	if err := checkinRepo.Create(checkin); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.auditLog.Log(0, "checkin", "checkin", checkin.ID, nil, checkin, "办理入住"); err != nil {
		log.Printf("Audit log failed for checkin %d: %v", checkin.ID, err)
	}
	return checkin, nil
}

func (s *CheckinService) FindByID(id uint) (*model.Checkin, error) {
	return s.repo.FindByID(id)
}

func (s *CheckinService) FindAll(page, pageSize int) ([]model.Checkin, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *CheckinService) Checkout(id uint) (*model.Checkin, error) {
	checkin, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if checkin.Status != model.CheckinStatusActive {
		return nil, nil
	}

	now := time.Now()
	checkin.CheckOutTime = &now
	checkin.Status = model.CheckinStatusCompleted

	if err := s.repo.Update(checkin); err != nil {
		return nil, err
	}

	room, err := s.roomRepo.FindByID(checkin.RoomID)
	if err != nil {
		log.Printf("Checkout: room %d not found for checkin %d: %v", checkin.RoomID, checkin.ID, err)
	} else {
		room.Status = model.RoomStatusVacant
		if err := s.roomRepo.Update(room); err != nil {
			log.Printf("Checkout: failed to update room %d status: %v", room.ID, err)
		}
	}

	if err := s.auditLog.Log(0, "checkout", "checkin", checkin.ID, nil, checkin, "办理签离"); err != nil {
		log.Printf("Audit log failed for checkout %d: %v", checkin.ID, err)
	}
	return checkin, nil
}

func (s *CheckinService) Delete(id uint) error {
	return s.repo.Delete(id)
}
