package service

import (
	"errors"

	"hotel-backend/internal/model"
	"hotel-backend/internal/repository"
	"gorm.io/gorm"
)

type AllocateRequirements struct {
	GuestCount int
	RoomType   string
	Floor      int
}

type AllocationStrategy interface {
	Select(rooms []model.Room, req AllocateRequirements) (*model.Room, error)
}

type RoomAllocator struct {
	roomRepo  *repository.RoomRepo
	orderRepo *repository.OrderRepo
	strategy  AllocationStrategy
}

func NewRoomAllocator(roomRepo *repository.RoomRepo, orderRepo *repository.OrderRepo, strategy AllocationStrategy) *RoomAllocator {
	return &RoomAllocator{
		roomRepo:  roomRepo,
		orderRepo: orderRepo,
		strategy:  strategy,
	}
}

func (a *RoomAllocator) Allocate(req AllocateRequirements, checkIn, checkOut string) (*model.Room, error) {
	return a.allocateWithRepos(a.roomRepo, a.orderRepo, req, checkIn, checkOut)
}

func (a *RoomAllocator) AllocateWithTx(tx *gorm.DB, req AllocateRequirements, checkIn, checkOut string) (*model.Room, error) {
	return a.allocateWithRepos(a.roomRepo.WithTx(tx), a.orderRepo.WithTx(tx), req, checkIn, checkOut)
}

func (a *RoomAllocator) allocateWithRepos(roomRepo *repository.RoomRepo, orderRepo *repository.OrderRepo, req AllocateRequirements, checkIn, checkOut string) (*model.Room, error) {
	filter := repository.RoomFilter{
		MinCapacity: req.GuestCount,
		Type:        req.RoomType,
		Floor:       req.Floor,
	}
	candidates, err := roomRepo.FindCandidates(filter)
	if err != nil {
		return nil, err
	}

	availableRooms, err := checkDateAvailability(orderRepo, candidates, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if len(availableRooms) == 0 {
		return nil, ErrNoRoomAvailable
	}

	best, err := a.strategy.Select(availableRooms, req)
	if err != nil {
		return nil, err
	}

	return roomRepo.FindByIDForUpdate(best.ID)
}

func checkDateAvailability(orderRepo *repository.OrderRepo, rooms []model.Room, checkIn, checkOut string) ([]model.Room, error) {
	if len(rooms) == 0 {
		return rooms, nil
	}

	roomIDs := make([]uint, len(rooms))
	for i, room := range rooms {
		roomIDs[i] = room.ID
	}

	overlappingIDs, err := orderRepo.FindOverlappingRoomIDs(roomIDs, checkIn, checkOut)
	if err != nil {
		return nil, err
	}

	blocked := make(map[uint]struct{}, len(overlappingIDs))
	for _, id := range overlappingIDs {
		blocked[id] = struct{}{}
	}

	var available []model.Room
	for _, room := range rooms {
		if _, ok := blocked[room.ID]; !ok {
			available = append(available, room)
		}
	}
	return available, nil
}

type FloorStrategy struct {
	ascending bool
}

func NewFloorStrategy(ascending bool) *FloorStrategy {
	return &FloorStrategy{ascending: ascending}
}

func (s *FloorStrategy) Select(rooms []model.Room, req AllocateRequirements) (*model.Room, error) {
	if len(rooms) == 0 {
		return nil, ErrNoRoomAvailable
	}

	best := &rooms[0]
	for i := 1; i < len(rooms); i++ {
		if s.ascending {
			if rooms[i].Floor < best.Floor {
				best = &rooms[i]
			}
		} else {
			if rooms[i].Floor > best.Floor {
				best = &rooms[i]
			}
		}
	}
	return best, nil
}

func NewStrategyFromConfig(strategyName string) AllocationStrategy {
	switch strategyName {
	case "high_floor":
		return NewFloorStrategy(false)
	case "low_floor":
		fallthrough
	default:
		return NewFloorStrategy(true)
	}
}

var ErrNoRoomAvailable = errors.New("no room available for the requested criteria")

var ErrRoomAlreadyAllocated = errors.New("room has already been allocated")

var ErrCheckinNotActive = errors.New("checkin is not active")
