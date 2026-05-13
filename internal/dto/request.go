// internal/dto/request.go
package dto

import "strconv"

type PaginationRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

func (p *PaginationRequest) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

func (p PaginationRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func ParsePagination(pageStr, sizeStr string) PaginationRequest {
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	p := PaginationRequest{Page: page, PageSize: size}
	p.Normalize()
	return p
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type CreateRoomRequest struct {
	RoomNumber    string  `json:"room_number" binding:"required"`
	Type          string  `json:"type"`
	Floor         int     `json:"floor"`
	PricePerNight float64 `json:"price_per_night"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

type UpdateRoomRequest struct {
	RoomNumber    string  `json:"room_number"`
	Type          string  `json:"type"`
	Floor         int     `json:"floor"`
	PricePerNight float64 `json:"price_per_night"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

type CreateOrderRequest struct {
	RoomID       uint    `json:"room_id" binding:"required"`
	CheckInDate  string  `json:"check_in_date" binding:"required"`
	CheckOutDate string  `json:"check_out_date" binding:"required"`
	TotalPrice   float64 `json:"total_price"`
}

type UpdateOrderRequest struct {
	CheckInDate  string  `json:"check_in_date"`
	CheckOutDate string  `json:"check_out_date"`
	TotalPrice   float64 `json:"total_price"`
	Status       string  `json:"status"`
}

type CreateCheckinRequest struct {
	OrderID              *uint  `json:"order_id"`
	UserID               uint   `json:"user_id" binding:"required"`
	RoomID               uint   `json:"room_id" binding:"required"`
	ExpectedCheckoutTime string `json:"expected_checkout_time" binding:"required"`
}
