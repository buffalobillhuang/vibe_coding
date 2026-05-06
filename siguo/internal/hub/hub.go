package hub

import (
	"errors"
	"strings"
	"sync"

	"siguo/internal/room"
)

type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]*room.Room
	maxRooms int
}

func New(maxRooms int) *Hub {
	if maxRooms <= 0 {
		maxRooms = 100
	}
	return &Hub{rooms: map[string]*room.Room{}, maxRooms: maxRooms}
}

func (h *Hub) Create() (*room.Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.rooms) >= h.maxRooms {
		return nil, errors.New("too many rooms")
	}
	for i := 0; i < 20; i++ {
		code := room.NewCode()
		if h.rooms[code] == nil {
			r := room.New(code)
			h.rooms[code] = r
			return r, nil
		}
	}
	return nil, errors.New("failed to allocate room code")
}

func (h *Hub) Get(code string) (*room.Room, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	r, ok := h.rooms[normalize(code)]
	return r, ok
}

func normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
