package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/obfuscate"

	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

var ErrOrderNotFoundInCancel = errors.New("order not found")
var ErrOrderNotBelongToUser = errors.New("order does not belong to the user")
var ErrOrderNotPending = errors.New("order is not in pending status")
var ErrOrderNotCancelRequested = errors.New("order is not in cancel_requested status")
var ErrRoomStatusConflict = errors.New("room status conflict during cancellation")
var ErrPastCheckIn = errors.New("cannot cancel an order past check-in date")

type OrderService struct {
	repo                *repository.OrderRepo
	roomRepo            *repository.RoomRepo
	userRepo            *repository.UserRepo
	allocator           *RoomAllocator
	auditLog            *AuditLogService
	notifSvc            *NotificationService
	db                  *gorm.DB
	cutoffHours         int
	defaultRejectReason string
	obfuscateKey        *hashids.HashID
}

func NewOrderService(repo *repository.OrderRepo, roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, allocator *RoomAllocator, auditLog *AuditLogService, notifSvc *NotificationService, db *gorm.DB, cutoffHours int, obfuscateKey *hashids.HashID, defaultRejectReason string) *OrderService {
	return &OrderService{repo: repo, roomRepo: roomRepo, userRepo: userRepo, allocator: allocator, auditLog: auditLog, notifSvc: notifSvc, db: db, cutoffHours: cutoffHours, obfuscateKey: obfuscateKey, defaultRejectReason: defaultRejectReason}
}

func (s *OrderService) DecodeCode(code string) (uint, error) {
	return obfuscate.Decode(s.obfuscateKey, code)
}

func (s *OrderService) encodeOrder(o *model.Order) {
	code, err := obfuscate.Encode(s.obfuscateKey, o.ID)
	if err != nil {
		log.Printf("Failed to encode order %d: %v", o.ID, err)
		return
	}
	o.OrderCode = code
}

func (s *OrderService) encodeOrders(orders []model.Order) {
	for i := range orders {
		s.encodeOrder(&orders[i])
	}
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
		allocatedRoom, err = roomRepo.FindByIDForUpdate(*roomID)
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

	ok, err := roomRepo.UpdateStatusIf(allocatedRoom.ID, model.RoomStatusVacant, model.RoomStatusReserved)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrRoomAlreadyAllocated
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.auditLog.Log(0, "created", "order", order.ID, nil, order, "创建订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}
	s.encodeOrder(order)
	return order, nil
}

func (s *OrderService) FindByID(id uint) (*model.Order, error) {
	order, err := s.repo.FindByID(id)
	if err == nil {
		s.encodeOrder(order)
	}
	return order, err
}

func (s *OrderService) FindAll(page, pageSize int) ([]model.Order, int64, error) {
	orders, total, err := s.repo.FindAll(page, pageSize)
	if err == nil {
		s.encodeOrders(orders)
	}
	return orders, total, err
}

func (s *OrderService) FindByUserID(userID uint, page, pageSize int) ([]model.Order, int64, error) {
	orders, total, err := s.repo.FindByUserID(userID, page, pageSize)
	if err == nil {
		s.encodeOrders(orders)
	}
	return orders, total, err
}

func (s *OrderService) Update(id uint, checkIn, checkOut, status string, price float64) (*model.Order, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.repo.WithTx(tx)
	roomRepo := s.roomRepo.WithTx(tx)

	order, err := orderRepo.FindByID(id)
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
		if oldOrder.Status == model.OrderStatusCancelRequested {
			if status != model.OrderStatusCancelled && status != model.OrderStatusPending {
				return nil, fmt.Errorf("invalid status transition from cancel_requested to %s", status)
			}
			if status == model.OrderStatusCancelled {
				ok, rErr := roomRepo.UpdateStatusIf(order.RoomID, model.RoomStatusReserved, model.RoomStatusVacant)
				if rErr != nil {
					return nil, rErr
				}
				if !ok {
					return nil, ErrRoomStatusConflict
				}
				if _, err := s.notifSvc.Create(order.UserID, "cancel_approved", "取消申请已通过",
					fmt.Sprintf("您的订单 #%d 取消申请已通过，房间已释放。", order.ID)); err != nil {
					log.Printf("Notification failed for order %d: %v", order.ID, err)
				}
			} else {
				if _, err := s.notifSvc.Create(order.UserID, "cancel_rejected", "取消申请已被驳回",
					fmt.Sprintf("您的订单 #%d 取消申请未通过，订单恢复正常。", order.ID)); err != nil {
					log.Printf("Notification failed for order %d: %v", order.ID, err)
				}
			}
		}
		order.Status = status
	}
	if price > 0 {
		order.TotalPrice = price
	}
	if err := orderRepo.Update(order); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
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
	s.encodeOrder(order)
	return order, nil
}

func (s *OrderService) RejectCancel(id uint, reason string) (*model.Order, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.repo.WithTx(tx)

	order, err := orderRepo.FindByIDForUpdate(tx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != model.OrderStatusCancelRequested {
		return nil, ErrOrderNotCancelRequested
	}

	order.Status = model.OrderStatusPending
	if err := orderRepo.Update(order); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if reason == "" {
		reason = s.defaultRejectReason
	}
	if _, err := s.notifSvc.Create(order.UserID, "cancel_rejected", "取消申请已被驳回",
		fmt.Sprintf("您的订单 #%d 取消申请未通过，原因：%s", order.ID, reason)); err != nil {
		log.Printf("Notification failed for order %d: %v", order.ID, err)
	}

	if err := s.auditLog.Log(0, "rejected_cancel", "order", order.ID, nil, order, "驳回取消申请"); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}

	s.encodeOrder(order)
	return order, nil
}

func (s *OrderService) Confirm(id uint) (*model.Order, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.repo.WithTx(tx)

	order, err := orderRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if order.Status != model.OrderStatusPending {
		return nil, ErrOrderNotPending
	}

	oldOrder := *order
	order.Status = model.OrderStatusConfirmed
	if err := orderRepo.Update(order); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.auditLog.Log(0, "confirmed", "order", order.ID, &oldOrder, order, "确认订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}
	s.encodeOrder(order)
	return order, nil
}

func (s *OrderService) Cancel(id, userID uint, reason string) (*model.Order, bool, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, false, tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.repo.WithTx(tx)
	roomRepo := s.roomRepo.WithTx(tx)

	order, err := orderRepo.FindByIDForUpdate(tx, id)
	if err != nil {
		return nil, false, ErrOrderNotFoundInCancel
	}
	if order.UserID != userID {
		return nil, false, ErrOrderNotBelongToUser
	}
	if order.Status != model.OrderStatusPending {
		return nil, false, ErrOrderNotPending
	}

	order.CancelReason = reason

	checkInTime, err := time.Parse("2006-01-02", order.CheckInDate)
	if err != nil {
		return nil, false, err
	}

	// 日期层面的检查：入住日期在今日之前 → 拒绝取消
	today := time.Now().Truncate(24 * time.Hour)
	if checkInTime.Before(today) {
		return nil, false, ErrPastCheckIn
	}

	// 时间层面的检查：距入住的实际小时数决定自动取消还是审核
	hoursUntilCheckIn := time.Until(checkInTime).Hours()
	autoCancel := hoursUntilCheckIn > float64(s.cutoffHours)

	if autoCancel {
		ok, err := roomRepo.UpdateStatusIf(order.RoomID, model.RoomStatusReserved, model.RoomStatusVacant)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, ErrRoomStatusConflict
		}
		order.Status = model.OrderStatusCancelled
	} else {
		order.Status = model.OrderStatusCancelRequested
	}

	if err := orderRepo.Update(order); err != nil {
		return nil, false, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, false, err
	}

	if err := s.auditLog.Log(userID, "cancelled", "order", order.ID, nil, order, "取消订单: "+reason); err != nil {
		log.Printf("Audit log failed for order %d: %v", order.ID, err)
	}

	if autoCancel {
		if _, err := s.notifSvc.Create(userID, "order_cancelled", "订单已取消",
			fmt.Sprintf("您的订单 #%d 已自动取消。原因：%s", order.ID, reason)); err != nil {
			log.Printf("Notification failed for order %d: %v", order.ID, err)
		}
	} else {
		s.notifyStaffCancelRequest(order)
	}

	s.encodeOrder(order)
	return order, autoCancel, nil
}

func (s *OrderService) FindCancelRequests(page, pageSize int) ([]model.Order, int64, error) {
	orders, total, err := s.repo.FindByStatus(model.OrderStatusCancelRequested, page, pageSize)
	if err == nil {
		s.encodeOrders(orders)
	}
	return orders, total, err
}

func (s *OrderService) notifyStaffCancelRequest(order *model.Order) {
	users, _ := s.userRepo.FindByRoles("employee", "admin")
	title := "取消订单审核"
	username := ""
	if order.User != nil {
		username = order.User.Username
	}
	content := fmt.Sprintf("用户 %s 的订单 #%d 已提交取消申请（理由：%s），请审核。", username, order.ID, order.CancelReason)
	for _, u := range users {
		if _, err := s.notifSvc.Create(u.ID, "cancel_request_alert", title, content); err != nil {
			log.Printf("Notification failed for user %d: %v", u.ID, err)
		}
	}
}

func (s *OrderService) Delete(id uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.repo.WithTx(tx)
	roomRepo := s.roomRepo.WithTx(tx)

	order, err := orderRepo.FindByIDForUpdate(tx, id)
	if err != nil {
		return err
	}

	// 非已取消状态需要释放房间
	if order.Status != model.OrderStatusCancelled {
		ok, err := roomRepo.UpdateStatusIf(order.RoomID, model.RoomStatusReserved, model.RoomStatusVacant)
		if err != nil {
			return err
		}
		if !ok {
			return ErrRoomStatusConflict
		}
	}

	if err := orderRepo.Delete(id); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	if err := s.auditLog.Log(0, "deleted", "order", id, order, nil, "删除订单"); err != nil {
		log.Printf("Audit log failed for order %d: %v", id, err)
	}
	return nil
}
