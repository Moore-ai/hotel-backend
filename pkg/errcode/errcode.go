package errcode

import "github.com/gin-gonic/gin"

const (
	Success          = 0
	ErrInternal      = 500
	ErrBadRequest    = 400
	ErrUnauthorized  = 401
	ErrForbidden     = 403
	ErrNotFound      = 404
	ErrConflict      = 409

	ErrInvalidCredentials = 1001
	ErrTokenExpired       = 1002
	ErrTokenInvalid       = 1003
	ErrUserNotFound       = 2001
	ErrUsernameDuplicate  = 2002
	ErrRoomNotFound       = 3001
	ErrRoomOccupied       = 3002
	ErrNoRoomAvailable    = 3003
	ErrOrderNotFound      = 4001
	ErrCheckinNotFound    = 5001
	ErrAlreadyCheckedOut  = 5002
)

var messages = map[int]string{
	Success:                "success",
	ErrInternal:            "internal server error",
	ErrBadRequest:          "bad request",
	ErrUnauthorized:        "unauthorized",
	ErrForbidden:           "forbidden",
	ErrNotFound:            "not found",
	ErrConflict:            "conflict",
	ErrInvalidCredentials:  "invalid username or password",
	ErrTokenExpired:        "token expired",
	ErrTokenInvalid:        "invalid token",
	ErrUserNotFound:        "user not found",
	ErrUsernameDuplicate:   "username already exists",
	ErrRoomNotFound:        "room not found",
	ErrRoomOccupied:        "room is currently occupied",
	ErrNoRoomAvailable:     "no room available for the requested criteria",
	ErrOrderNotFound:       "order not found",
	ErrCheckinNotFound:     "checkin not found",
	ErrAlreadyCheckedOut:   "already checked out",
}

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "unknown error"
}

func Write(c *gin.Context, code int) {
	httpStatus := 200
	switch code {
	case ErrBadRequest:
		httpStatus = 400
	case ErrUnauthorized:
		httpStatus = 401
	case ErrForbidden:
		httpStatus = 403
	case ErrNotFound:
		httpStatus = 404
	case ErrInternal:
		httpStatus = 500
	}
	c.JSON(httpStatus, gin.H{"code": code, "message": Message(code)})
}
