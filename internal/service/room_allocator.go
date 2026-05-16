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
	rooms, err := a.roomRepo.FindAllRooms()
	if err != nil {
		return nil, err
	}

	candidates := filterCandidates(rooms, req)
	availableRooms, err := a.checkDateAvailability(candidates, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if len(availableRooms) == 0 {
		return nil, ErrNoRoomAvailable
	}

	return a.strategy.Select(availableRooms, req)
}

func filterCandidates(rooms []model.Room, req AllocateRequirements) []model.Room {
	var candidates []model.Room
	for _, room := range rooms {
		if req.GuestCount > 0 && room.Capacity < req.GuestCount {
			continue
		}
		if req.RoomType != "" && room.Type != req.RoomType {
			continue
		}
		if req.Floor > 0 && room.Floor != req.Floor {
			continue
		}
		candidates = append(candidates, room)
	}
	return candidates
}

func (a *RoomAllocator) checkDateAvailability(rooms []model.Room, checkIn, checkOut string) ([]model.Room, error) {
	if len(rooms) == 0 {
		return rooms, nil
	}

	roomIDs := make([]uint, len(rooms))
	for i, room := range rooms {
		roomIDs[i] = room.ID
	}

	overlappingIDs, err := a.orderRepo.FindOverlappingRoomIDs(roomIDs, checkIn, checkOut)
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
