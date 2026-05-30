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

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type CreateGuestRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type UpdateGuestRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type CreateEmployeeRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=64"`
	Password string  `json:"password" binding:"required,min=6"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	HireDate string  `json:"hire_date"`
	Salary   float64 `json:"salary"`
	Notes    string  `json:"notes"`
}

type UpdateEmployeeRequest struct {
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	HireDate string  `json:"hire_date"`
	Salary   float64 `json:"salary"`
	Notes    string  `json:"notes"`
}

type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type UpdateAdminRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type CreateRoomRequest struct {
	RoomNumber    string  `json:"room_number" binding:"required"`
	Type          string  `json:"type"`
	Capacity      int     `json:"capacity"`
	Floor         int     `json:"floor"`
	PricePerNight float64 `json:"price_per_night"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

type UpdateRoomRequest struct {
	RoomNumber    string  `json:"room_number"`
	Type          string  `json:"type"`
	Capacity      int     `json:"capacity"`
	Floor         int     `json:"floor"`
	PricePerNight float64 `json:"price_per_night"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
}

type CreateOrderRequest struct {
	RoomID             uint    `json:"room_id"`
	GuestCount         int     `json:"guest_count"`
	RoomTypePreference string  `json:"room_type_preference"`
	CheckInDate        string  `json:"check_in_date" binding:"required"`
	CheckOutDate       string  `json:"check_out_date" binding:"required"`
	TotalPrice         float64 `json:"total_price"`
}

type UpdateOrderRequest struct {
	CheckInDate  string  `json:"check_in_date"`
	CheckOutDate string  `json:"check_out_date"`
	TotalPrice   float64 `json:"total_price"`
	Status       string  `json:"status"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=2"`
}

type RejectCancelRequest struct {
	Reason string `json:"reason"`
}

type CreateWaiterRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=64"`
	Password string  `json:"password" binding:"required,min=6"`
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	HireDate string  `json:"hire_date"`
	Salary   float64 `json:"salary"`
	Notes    string  `json:"notes"`
}

type UpdateWaiterRequest struct {
	Name     string  `json:"name"`
	Phone    string  `json:"phone"`
	Email    string  `json:"email"`
	HireDate string  `json:"hire_date"`
	Salary   float64 `json:"salary"`
	Notes    string  `json:"notes"`
}

type ServiceRequest struct {
	Content string `json:"content" binding:"required"`
	Note    string `json:"note"`
}

type CreateCheckinRequest struct {
	OrderID              *uint  `json:"order_id"`
	UserCode             string `json:"user_code" binding:"required"`
	RoomID               uint   `json:"room_id"`
	ExpectedCheckoutTime string `json:"expected_checkout_time" binding:"required"`
}

type CreateAppealRequest struct {
	Reason string `json:"reason" binding:"required,min=2"`
}

type ReviewAppealRequest struct {
	Action     string `json:"action" binding:"required"` // approved | rejected
	ReviewNote string `json:"review_note"`
}

type ChatRequest struct {
	Message        string `json:"message" binding:"required"`
	ConversationID string `json:"conversation_id"`
}
