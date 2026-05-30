package service

import (
	"errors"

	"hotel-backend/config"
	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"hotel-backend/pkg/hash"
	"hotel-backend/pkg/jwt"
)

type AuthService struct {
	userRepo     *repository.UserRepo
	guestRepo    *repository.GuestRepo
	employeeRepo *repository.EmployeeRepo
	adminRepo    *repository.AdminRepo
	waiterRepo   *repository.WaiterRepo
	userService  *UserService
	jwtCfg       config.JWTConfig
}

func NewAuthService(userRepo *repository.UserRepo, guestRepo *repository.GuestRepo, employeeRepo *repository.EmployeeRepo, adminRepo *repository.AdminRepo, waiterRepo *repository.WaiterRepo, userService *UserService, jwtCfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		guestRepo:    guestRepo,
		employeeRepo: employeeRepo,
		adminRepo:    adminRepo,
		waiterRepo:   waiterRepo,
		userService:  userService,
		jwtCfg:       jwtCfg,
	}
}

func (s *AuthService) Login(username, password string, allowedRoles ...string) (string, string, int64, *model.User, string, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", "", 0, nil, "", ErrInvalidCredentials
	}

	if !hash.CheckPassword(user.PasswordHash, password) {
		return "", "", 0, nil, "", ErrInvalidCredentials
	}

	if len(allowedRoles) > 0 {
		roleAllowed := false
		for _, r := range allowedRoles {
			if user.Role == r {
				roleAllowed = true
				break
			}
		}
		if !roleAllowed {
			return "", "", 0, nil, "", ErrInvalidCredentials
		}
	}

	name := s.lookupProfileName(user.ID, user.Role)

	accessExp := s.jwtCfg.AccessTokenExpiry
	refreshExp := s.jwtCfg.RefreshTokenExpiry
	accessToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, accessExp)
	if err != nil {
		return "", "", 0, nil, "", err
	}
	refreshToken, err := jwt.GenerateToken(user.ID, user.Role, s.jwtCfg.Secret, refreshExp)
	if err != nil {
		return "", "", 0, nil, "", err
	}

	s.userService.encodeUser(user)
	return accessToken, refreshToken, int64(accessExp.Seconds()), user, name, nil
}

func (s *AuthService) lookupProfileName(userID uint, role string) string {
	switch role {
	case "guest":
		g, err := s.guestRepo.FindByUserID(userID)
		if err == nil {
			return g.Name
		}
	case "employee":
		e, err := s.employeeRepo.FindByUserID(userID)
		if err == nil {
			return e.Name
		}
	case "admin":
		a, err := s.adminRepo.FindByUserID(userID)
		if err == nil {
			return a.Name
		}
	case "waiter":
		w, err := s.waiterRepo.FindByUserID(userID)
		if err == nil {
			return w.Name
		}
	}
	return ""
}

func (s *AuthService) Logout() error {
	return nil
}

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already taken")
)

func (s *AuthService) Register(username, password, name, phone, email string) (*model.User, error) {
	return s.userService.Create(username, password, "guest", name, phone, email, "", 0, "")
}

func (s *AuthService) DeleteAccount(userID uint) error {
	return s.userService.Delete(userID)
}
