package hub

import (
	"encoding/json"
	"testing"

	"siguo/internal/protocol"
	"siguo/internal/room"
)

func TestCreateAllowsLobbyRoomsAtActiveCap(t *testing.T) {
	h := New(1)

	lobby, err := h.Create()
	if err != nil {
		t.Fatalf("Create() lobby error = %v", err)
	}
	if _, err := h.Create(); err != nil {
		t.Fatalf("Create() should allow extra lobby rooms when no game is active: %v", err)
	}

	startRoom(t, lobby)

	next, err := h.Create()
	if err != nil {
		t.Fatalf("Create() should allow lobby rooms when active games are capped: %v", err)
	}
	startRoom(t, next)
	if next.Active() {
		t.Fatal("second room should remain in lobby when active room cap is full")
	}
}

func TestNewUsesTwentyFiveActiveRoomsByDefault(t *testing.T) {
	h := New(0)
	if h.maxRooms != 25 {
		t.Fatalf("default maxRooms = %d, want 25", h.maxRooms)
	}
}

func startRoom(t *testing.T, r *room.Room) {
	t.Helper()
	host, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(host) error = %v", err)
	}
	for _, name := range []string{"east", "south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	raw, err := json.Marshal(protocol.ClientMessage{Type: "room.start", Seq: 1})
	if err != nil {
		t.Fatalf("Marshal(room.start) error = %v", err)
	}
	r.Handle(host.Token, raw)
}
