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
	ErrRoomNotFound       = 3001
	ErrRoomOccupied       = 3002
	ErrNoRoomAvailable    = 3003
	ErrNoWaiterAvailable  = 3004
	ErrOrderNotFound      = 4001
	ErrOrderNotPending    = 4002
	ErrAppealNotFound      = 4003
	ErrAppealNotPending    = 4004
	ErrAppealExists        = 4005
	ErrCheckinNotFound    = 5001
	ErrAlreadyCheckedOut  = 5002
	ErrLLMUnavailable     = 6001
	ErrLLMInvalidResp     = 6002
	ErrRateLimited        = 6003
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
	ErrRoomNotFound:        "room not found",
	ErrRoomOccupied:        "room is currently occupied",
	ErrNoRoomAvailable:     "no room available for the requested criteria",
	ErrNoWaiterAvailable:   "no waiter available",
	ErrOrderNotFound:       "order not found",
	ErrOrderNotPending:     "order status is not pending, cannot cancel",
		ErrAppealNotFound:      "appeal not found",
		ErrAppealNotPending:    "appeal is not in pending status",
		ErrAppealExists:        "an active appeal already exists for this order",
	ErrCheckinNotFound:     "checkin not found",
	ErrAlreadyCheckedOut:   "already checked out",
	ErrLLMUnavailable:     "AI 管家暂时不可用，请稍后再试",
	ErrLLMInvalidResp:     "AI 响应解析失败，请重试",
	ErrRateLimited:        "请求过于频繁，请稍后再试",
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
	case ErrRateLimited:
		httpStatus = 429
	}
	c.JSON(httpStatus, gin.H{"code": code, "message": Message(code)})
}
