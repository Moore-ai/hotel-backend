package service

import (
	"errors"
	"fmt"
	"log"
	"math/rand"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/obfuscate"

	"github.com/speps/go-hashids/v2"
	"gorm.io/gorm"
)

const (
	AppealReviewStrategyAdminOnly = "admin_only"
	AppealReviewStrategyRandomOne = "random_one"
)

var (
	ErrAppealNotFound      = errors.New("appeal not found")
	ErrAppealNotPending    = errors.New("appeal is not in pending status")
	ErrAppealExists        = errors.New("an active appeal already exists for this order")
	ErrInvalidAppealAction = errors.New("invalid appeal action, must be approved or rejected")
)

type AppealService struct {
	repo           *repository.AppealRepo
	orderRepo      *repository.OrderRepo
	roomRepo       *repository.RoomRepo
	userRepo       *repository.UserRepo
	notifSvc       *NotificationService
	auditLog       *AuditLogService
	db             *gorm.DB
	reviewStrategy string
	reviewStaffIDs []uint
	obfuscateKey   *hashids.HashID
}

func NewAppealService(repo *repository.AppealRepo, orderRepo *repository.OrderRepo, roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, notifSvc *NotificationService, auditLog *AuditLogService, db *gorm.DB, reviewStrategy string, reviewStaffIDs []uint, obfuscateKey *hashids.HashID) *AppealService {
	return &AppealService{repo: repo, orderRepo: orderRepo, roomRepo: roomRepo, userRepo: userRepo, notifSvc: notifSvc, auditLog: auditLog, db: db, reviewStrategy: reviewStrategy, reviewStaffIDs: reviewStaffIDs, obfuscateKey: obfuscateKey}
}

func (s *AppealService) DecodeCode(code string) (uint, error) {
	return obfuscate.Decode(s.obfuscateKey, code)
}

func (s *AppealService) encodeAppeal(a *model.Appeal) {
	code, err := obfuscate.Encode(s.obfuscateKey, a.ID)
	if err != nil {
		log.Printf("Failed to encode appeal %d: %v", a.ID, err)
		return
	}
	a.AppealCode = code
}

func (s *AppealService) encodeAppeals(appeals []model.Appeal) {
	for i := range appeals {
		s.encodeAppeal(&appeals[i])
	}
}

func (s *AppealService) Create(orderID, userID uint, reason string) (*model.Appeal, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	orderRepo := s.orderRepo.WithTx(tx)
	appealRepo := s.repo.WithTx(tx)

	order, err := orderRepo.FindByIDForUpdate(tx, orderID)
	if err != nil {
		return nil, ErrOrderNotFoundInCancel
	}
	if order.UserID != userID {
		return nil, ErrOrderNotBelongToUser
	}
	if order.Status != model.OrderStatusPending {
		return nil, ErrOrderNotPending
	}

	existing, err := appealRepo.FindByOrderID(orderID)
	if err == nil {
		if existing.Status == model.AppealStatusPending {
			return nil, ErrAppealExists
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	appeal := &model.Appeal{
		OrderID: orderID,
		UserID:  userID,
		Reason:  reason,
		Status:  model.AppealStatusPending,
	}
	if err := appealRepo.Create(appeal); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.notifyReviewers(appeal)

	if err := s.auditLog.Log(userID, "appealed", "order", orderID, nil, order, "申诉订单: "+reason); err != nil {
		log.Printf("Audit log failed for appeal order %d: %v", orderID, err)
	}

	s.encodeAppeal(appeal)
	return appeal, nil
}

func (s *AppealService) FindByID(id uint) (*model.Appeal, error) {
	appeal, err := s.repo.FindByID(id)
	if err == nil {
		s.encodeAppeal(appeal)
	}
	return appeal, err
}

func (s *AppealService) FindAll(status string, page, pageSize int) ([]model.Appeal, int64, error) {
	appeals, total, err := s.repo.FindByStatus(status, page, pageSize)
	if err == nil {
		s.encodeAppeals(appeals)
	}
	return appeals, total, err
}

func (s *AppealService) Review(id, reviewerID uint, action, reviewNote string) (*model.Appeal, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	appealRepo := s.repo.WithTx(tx)
	orderRepo := s.orderRepo.WithTx(tx)
	roomRepo := s.roomRepo.WithTx(tx)

	appeal, err := appealRepo.FindByIDForUpdate(tx, id)
	if err != nil {
		return nil, ErrAppealNotFound
	}
	if appeal.Status != model.AppealStatusPending {
		return nil, ErrAppealNotPending
	}

	appeal.ReviewerID = &reviewerID
	if reviewNote != "" {
		appeal.ReviewNote = &reviewNote
	}

	var order *model.Order

	switch action {
	case model.AppealStatusApproved:
		ord, oErr := orderRepo.FindByIDForUpdate(tx, appeal.OrderID)
		if oErr != nil {
			return nil, oErr
		}
		order = ord
		ok, rErr := roomRepo.UpdateStatusIf(order.RoomID, model.RoomStatusReserved, model.RoomStatusVacant)
		if rErr != nil {
			return nil, rErr
		}
		if !ok {
			return nil, ErrRoomStatusConflict
		}
		order.Status = model.OrderStatusCancelled
		if err := orderRepo.Update(order); err != nil {
			return nil, err
		}
		appeal.Status = model.AppealStatusApproved
	case model.AppealStatusRejected:
		appeal.Status = model.AppealStatusRejected
	default:
		return nil, ErrInvalidAppealAction
	}

	if err := appealRepo.Update(appeal); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	notifType := "appeal_rejected"
	notifTitle := "申诉已被驳回"
	notifContent := fmt.Sprintf("您的订单 #%d 的申诉未通过，原因：%s", appeal.OrderID, reviewNote)
	auditAction := "appeal_rejected"
	auditDetail := "申诉驳回: " + reviewNote
	if action == model.AppealStatusApproved {
		notifType = "appeal_approved"
		notifTitle = "申诉已通过"
		notifContent = fmt.Sprintf("您的订单 #%d 的申诉已通过，订单已取消。", appeal.OrderID)
		auditAction = "appeal_approved"
		auditDetail = "申诉通过"
	}
	if _, err := s.notifSvc.Create(appeal.UserID, notifType, notifTitle, notifContent); err != nil {
		log.Printf("Notification failed for appeal user %d: %v", appeal.UserID, err)
	}
	if err := s.auditLog.Log(reviewerID, auditAction, "order", appeal.OrderID, nil, order, auditDetail); err != nil {
		log.Printf("Audit log failed: %v", err)
	}

	s.encodeAppeal(appeal)
	return appeal, nil
}

func (s *AppealService) notifyReviewers(appeal *model.Appeal) {
	title := fmt.Sprintf("订单 %d 申诉审核", appeal.OrderID)

	var users []model.User
	switch s.reviewStrategy {
	case AppealReviewStrategyRandomOne:
		if len(s.reviewStaffIDs) > 0 {
			for _, id := range s.reviewStaffIDs {
				u, err := s.userRepo.FindByID(id)
				if err != nil {
					log.Printf("Review staff id %d not found: %v", id, err)
					continue
				}
				users = append(users, *u)
			}
		}
		if len(users) == 0 {
			all, _ := s.userRepo.FindByRole("employee")
			users = all
		}
		if len(users) > 0 {
			chosen := users[rand.Intn(len(users))]
			users = []model.User{chosen}
		}
	default:
		users, _ = s.userRepo.FindByRole("admin")
	}

	if len(users) == 0 {
		log.Printf("no reviewer available to notify for appeal of order #%d", appeal.OrderID)
		return
	}

	for _, u := range users {
		if _, err := s.notifSvc.Create(u.ID, "appeal_request", title, appeal.Reason); err != nil {
			log.Printf("Notification failed for user %d: %v", u.ID, err)
		}
	}
}
