package service

import (
	"errors"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/hash"

	"gorm.io/gorm"
)

type UserService struct {
	repo         *repository.UserRepo
	guestRepo    *repository.GuestRepo
	employeeRepo *repository.EmployeeRepo
	adminRepo    *repository.AdminRepo
	waiterRepo   *repository.WaiterRepo
	auditLog     *AuditLogService
	db           *gorm.DB
}

func NewUserService(repo *repository.UserRepo, guestRepo *repository.GuestRepo, employeeRepo *repository.EmployeeRepo, adminRepo *repository.AdminRepo, waiterRepo *repository.WaiterRepo, auditLog *AuditLogService, db *gorm.DB) *UserService {
	return &UserService{repo: repo, guestRepo: guestRepo, employeeRepo: employeeRepo, adminRepo: adminRepo, waiterRepo: waiterRepo, auditLog: auditLog, db: db}
}

var ErrUsernameExists = errors.New("username already exists")

func (s *UserService) Create(username, password, role, name, phone, email string) (*model.User, error) {
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return nil, ErrUsernameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	pw, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer tx.Rollback()

	userRepo := s.repo.WithTx(tx)

	user := &model.User{
		Username:     username,
		PasswordHash: pw,
		Role:         role,
	}
	if err := userRepo.Create(user); err != nil {
		return nil, err
	}

	switch role {
	case "guest":
		guestRepo := s.guestRepo.WithTx(tx)
		if err := guestRepo.Create(&model.Guest{
			UserID: user.ID,
			Name:   name,
			Phone:  phone,
			Email:  email,
		}); err != nil {
			return nil, err
		}
	case "employee":
		empRepo := s.employeeRepo.WithTx(tx)
		if err := empRepo.Create(&model.Employee{
			UserID: user.ID,
			Name:   name,
			Phone:  phone,
			Email:  email,
		}); err != nil {
			return nil, err
		}
	case "admin":
		adminRepo := s.adminRepo.WithTx(tx)
		if err := adminRepo.Create(&model.Admin{
			UserID: user.ID,
			Name:   name,
			Phone:  phone,
			Email:  email,
		}); err != nil {
			return nil, err
		}
	case "waiter":
		waiterRepo := s.waiterRepo.WithTx(tx)
		if err := waiterRepo.Create(&model.Waiter{
			UserID: user.ID,
			Name:   name,
			Phone:  phone,
			Email:  email,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.auditLog.Log(0, "created", "user", user.ID, nil, user, "创建用户 "+user.Username)
	return user, nil
}

func (s *UserService) FindByID(id uint) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) FindAll(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *UserService) Update(id uint, username, password, role, name, phone, email string) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldUser := *user
	if username != "" {
		user.Username = username
	}
	if password != "" {
		pw, err := hash.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = pw
	}
	if role != "" {
		user.Role = role
	}
	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	if name != "" || phone != "" || email != "" {
		switch user.Role {
		case "guest":
			g, err := s.guestRepo.FindByUserID(user.ID)
			if err == nil {
				if name != "" {
					g.Name = name
				}
				if phone != "" {
					g.Phone = phone
				}
				if email != "" {
					g.Email = email
				}
				if err := s.guestRepo.Update(g); err != nil {
					return nil, err
				}
			}
		case "employee":
			e, err := s.employeeRepo.FindByUserID(user.ID)
			if err == nil {
				if name != "" {
					e.Name = name
				}
				if phone != "" {
					e.Phone = phone
				}
				if email != "" {
					e.Email = email
				}
				if err := s.employeeRepo.Update(e); err != nil {
					return nil, err
				}
			}
		case "admin":
			a, err := s.adminRepo.FindByUserID(user.ID)
			if err == nil {
				if name != "" {
					a.Name = name
				}
				if phone != "" {
					a.Phone = phone
				}
				if email != "" {
					a.Email = email
				}
				if err := s.adminRepo.Update(a); err != nil {
					return nil, err
				}
			}
		case "waiter":
			w, err := s.waiterRepo.FindByUserID(user.ID)
			if err == nil {
				if name != "" {
					w.Name = name
				}
				if phone != "" {
					w.Phone = phone
				}
				if email != "" {
					w.Email = email
				}
				if err := s.waiterRepo.Update(w); err != nil {
					return nil, err
				}
			}
		}
	}

	s.auditLog.Log(0, "updated", "user", user.ID, &oldUser, user, "修改用户 "+user.Username)
	return user, nil
}

func (s *UserService) Delete(id uint) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	if err := s.deleteProfile(tx, user.Role, user.ID); err != nil {
		return err
	}

	userRepo := s.repo.WithTx(tx)
	if err := userRepo.Delete(user.ID); err != nil {
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	s.auditLog.Log(0, "deleted", "user", id, user, nil, "删除用户")
	return nil
}

func (s *UserService) deleteProfile(tx *gorm.DB, role string, userID uint) error {
	switch role {
	case "guest":
		g, err := s.guestRepo.FindByUserID(userID)
		if err != nil {
			return nil
		}
		return s.guestRepo.WithTx(tx).Delete(g.ID)
	case "employee":
		e, err := s.employeeRepo.FindByUserID(userID)
		if err != nil {
			return nil
		}
		return s.employeeRepo.WithTx(tx).Delete(e.ID)
	case "admin":
		a, err := s.adminRepo.FindByUserID(userID)
		if err != nil {
			return nil
		}
		return s.adminRepo.WithTx(tx).Delete(a.ID)
	case "waiter":
		w, err := s.waiterRepo.FindByUserID(userID)
		if err != nil {
			return nil
		}
		return s.waiterRepo.WithTx(tx).Delete(w.ID)
	}
	return nil
}
