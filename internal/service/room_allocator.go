package service

import (
	"errors"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
)

type AllocateRequirements struct {
	GuestCount int
	RoomType   string
	Floor      int
}

type RoomAllocator struct {
	roomRepo  *repository.RoomRepo
	orderRepo *repository.OrderRepo
}

func NewRoomAllocator(roomRepo *repository.RoomRepo, orderRepo *repository.OrderRepo) *RoomAllocator {
	return &RoomAllocator{roomRepo: roomRepo, orderRepo: orderRepo}
}

func (a *RoomAllocator) Allocate(req AllocateRequirements, checkIn, checkOut string) (*model.Room, error) {
	return nil, errors.New("room allocation algorithm not yet implemented")
}
