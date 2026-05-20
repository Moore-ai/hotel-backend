package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"hotel-backend/internal/dto"
	"hotel-backend/internal/service"
	"hotel-backend/pkg/errcode"
)

type EmployeeHandler struct {
	employeeService *service.EmployeeService
	userService     *service.UserService
}

func NewEmployeeHandler(employeeService *service.EmployeeService, userService *service.UserService) *EmployeeHandler {
	return &EmployeeHandler{employeeService: employeeService, userService: userService}
}

func (h *EmployeeHandler) List(c *gin.Context) {
	p := dto.ParsePagination(c.Query("page"), c.Query("page_size"))
	employees, total, err := h.employeeService.FindAll(p.Page, p.PageSize)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, dto.PageData{List: employees, Total: total, Page: p.Page, PageSize: p.PageSize})
}

func (h *EmployeeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	employee, err := h.employeeService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrNotFound)
		return
	}
	dto.Success(c, employee)
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req dto.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	_, err := h.userService.Create(req.Username, req.Password, "employee", req.Name, req.Phone, req.Email, req.HireDate, req.Salary, req.Notes)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Error(c, errcode.ErrBadRequest)
		return
	}
	employee, err := h.employeeService.Update(uint(id), req.Name, req.Phone, req.Email, req.HireDate, req.Salary, req.Notes)
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, employee)
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	employee, err := h.employeeService.FindByID(uint(id))
	if err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	if err := h.userService.Delete(employee.UserID); err != nil {
		dto.Error(c, errcode.ErrInternal)
		return
	}
	dto.Success(c, nil)
}
