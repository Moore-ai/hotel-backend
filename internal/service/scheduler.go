package service

import (
	"fmt"
	"log"
	"time"

	"hotel-backend/config"
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type Scheduler struct {
	checkinRepo    *repository.CheckinRepo
	scheduleRepo   *repository.CheckoutScheduleRepo
	notifService   *NotificationService
	checkinService *CheckinService
	userRepo       *repository.UserRepo
	cfg            config.CheckoutConfig
}

func NewScheduler(
	checkinRepo *repository.CheckinRepo,
	scheduleRepo *repository.CheckoutScheduleRepo,
	notifService *NotificationService,
	checkinService *CheckinService,
	userRepo *repository.UserRepo,
	cfg config.CheckoutConfig,
) *Scheduler {
	return &Scheduler{
		checkinRepo:    checkinRepo,
		scheduleRepo:   scheduleRepo,
		notifService:   notifService,
		checkinService: checkinService,
		userRepo:       userRepo,
		cfg:            cfg,
	}
}

func (s *Scheduler) Run() {
	interval := time.Duration(s.cfg.SchedulerInterval) * time.Second
	log.Printf("Scheduler started, checking every %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.tick()
	}
}

func (s *Scheduler) tick() {
	checkins, err := s.checkinRepo.FindActive()
	if err != nil {
		log.Printf("Scheduler: failed to find active checkins: %v", err)
		return
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	loc := now.Location()

	firstThreshold, _ := time.ParseInLocation("2006-01-02 15:04", today+" "+s.cfg.FirstThreshold, loc)
	secondThreshold, _ := time.ParseInLocation("2006-01-02 15:04", today+" "+s.cfg.SecondThreshold, loc)

	for _, c := range checkins {
		if c.ExpectedCheckoutTime.After(now) {
			continue
		}

		schedule, err := s.scheduleRepo.FindByCheckinID(c.ID)
		if err != nil {
			schedule = &model.CheckoutSchedule{
				CheckinID:       c.ID,
				FirstThreshold:  s.cfg.FirstThreshold,
				SecondThreshold: s.cfg.SecondThreshold,
			}
			_ = s.scheduleRepo.Create(schedule)
		}

		if now.After(secondThreshold) && !schedule.SecondNotified {
			s.notifyStaffAndAdmin(c)
			schedule.SecondNotified = true
			_ = s.scheduleRepo.Update(schedule)
			continue
		}

		if now.After(firstThreshold) && !schedule.FirstNotified {
			s.notifyGuest(c)
			schedule.FirstNotified = true
			_ = s.scheduleRepo.Update(schedule)
		}
	}
}

func (s *Scheduler) notifyGuest(c model.Checkin) {
	title := "签离提醒"
	content := fmt.Sprintf("您的 %s 号房应在 %s 前签离，请及时办理。", c.Room.RoomNumber, c.ExpectedCheckoutTime.Format("15:04"))
	s.notifService.Create(c.UserID, "checkout_reminder", title, content)
	log.Printf("Scheduler: sent guest notification to user %d for checkin %d", c.UserID, c.ID)
}

func (s *Scheduler) notifyStaffAndAdmin(c model.Checkin) {
	title := "超时签离告警"
	content := fmt.Sprintf("住户 %s 在 %s 号房超时未签离，应签离时间 %s，请跟进。", c.User.Name, c.Room.RoomNumber, c.ExpectedCheckoutTime.Format("2006-01-02 15:04"))

	employees, _ := s.userRepo.FindByRole("employee")
	admins, _ := s.userRepo.FindByRole("admin")

	for _, u := range employees {
		s.notifService.Create(u.ID, "overdue_alert", title, content)
	}
	for _, u := range admins {
		s.notifService.Create(u.ID, "overdue_alert", title, content)
	}
	log.Printf("Scheduler: sent staff/admin alerts for checkin %d", c.ID)
}
