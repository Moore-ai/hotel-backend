package repository

import (
	"hotel-backend/internal/model"

	"gorm.io/gorm"
)

type EmployeeRepo struct {
	db *gorm.DB
}

func NewEmployeeRepo(db *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

func (r *EmployeeRepo) Create(employee *model.Employee) error {
	return r.db.Create(employee).Error
}

func (r *EmployeeRepo) FindByID(id uint) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Preload("User").First(&employee, id).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *EmployeeRepo) FindByUserID(userID uint) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Where("user_id = ?", userID).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *EmployeeRepo) FindAll(page, pageSize int) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64
	r.db.Model(&model.Employee{}).Count(&total)
	err := r.db.Preload("User").Offset((page - 1) * pageSize).Limit(pageSize).Find(&employees).Error
	return employees, total, err
}

func (r *EmployeeRepo) Update(employee *model.Employee) error {
	return r.db.Save(employee).Error
}

func (r *EmployeeRepo) Delete(id uint) error {
	return r.db.Delete(&model.Employee{}, id).Error
}

func (r *EmployeeRepo) WithTx(tx *gorm.DB) *EmployeeRepo {
	return &EmployeeRepo{db: tx}
}
