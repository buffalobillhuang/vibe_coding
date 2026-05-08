package hub

import (
	"errors"
	"strings"
	"sync"

	"siguo/internal/protocol"
	"siguo/internal/room"
)

const defaultMaxActiveRooms = 25

type Hub struct {
	mu          sync.RWMutex
	rooms       map[string]*room.Room
	maxRooms    int
	activeRooms int
}

func New(maxRooms int) *Hub {
	if maxRooms <= 0 {
		maxRooms = defaultMaxActiveRooms
	}
	return &Hub{rooms: map[string]*room.Room{}, maxRooms: maxRooms}
}

func (h *Hub) Create() (*room.Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := 0; i < 20; i++ {
		code := room.NewCode()
		if h.rooms[code] == nil {
			r := room.New(code)
			r.SetActiveHooks(h.tryActivate, h.deactivate)
			h.rooms[code] = r
			return r, nil
		}
	}
	return nil, errors.New("failed to allocate room code")
}

func (h *Hub) ActiveRooms() []protocol.ActiveRoomInfo {
	h.mu.RLock()
	rooms := make([]*room.Room, 0, len(h.rooms))
	for _, r := range h.rooms {
		rooms = append(rooms, r)
	}
	h.mu.RUnlock()

	active := make([]protocol.ActiveRoomInfo, 0, len(rooms))
	for _, r := range rooms {
		if info, ok := r.ActiveSummary(); ok {
			active = append(active, info)
		}
	}
	return active
}

func (h *Hub) tryActivate() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.activeRooms >= h.maxRooms {
		return false
	}
	h.activeRooms++
	return true
}

func (h *Hub) deactivate() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.activeRooms > 0 {
		h.activeRooms--
	}
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
