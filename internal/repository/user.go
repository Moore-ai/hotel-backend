// internal/repository/user.go
package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindAll(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	r.db.Model(&model.User{}).Count(&total)
	err := r.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *UserRepo) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepo) FindByRole(role string) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

func (r *UserRepo) FindByRoles(roles ...string) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("role IN ?", roles).Find(&users).Error
	return users, err
}

func (r *UserRepo) WithTx(tx *gorm.DB) *UserRepo {
	return &UserRepo{db: tx}
}
