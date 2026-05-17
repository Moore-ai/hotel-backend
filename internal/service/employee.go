package service

import (
	"errors"
	"log"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type EmployeeService struct {
	repo     *repository.EmployeeRepo
	auditLog *AuditLogService
}

func NewEmployeeService(repo *repository.EmployeeRepo, auditLog *AuditLogService) *EmployeeService {
	return &EmployeeService{repo: repo, auditLog: auditLog}
}

func (s *EmployeeService) FindByID(id uint) (*model.Employee, error) {
	return s.repo.FindByID(id)
}

func (s *EmployeeService) FindAll(page, pageSize int) ([]model.Employee, int64, error) {
	return s.repo.FindAll(page, pageSize)
}

func (s *EmployeeService) Update(id uint, name, phone, email string) (*model.Employee, error) {
	employee, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	oldEmployee := *employee
	if name != "" {
		employee.Name = name
	}
	if phone != "" {
		employee.Phone = phone
	}
	if email != "" {
		employee.Email = email
	}
	if err := s.repo.Update(employee); err != nil {
		return nil, err
	}
	if err := s.auditLog.Log(0, "updated", "employee", employee.ID, &oldEmployee, employee, "修改员工"); err != nil {
		log.Printf("Audit log failed for employee %d: %v", employee.ID, err)
	}
	return employee, nil
}

var ErrEmployeeNotFound = errors.New("employee not found")
