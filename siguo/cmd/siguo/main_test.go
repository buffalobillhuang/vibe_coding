package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"siguo/internal/hub"
	"siguo/internal/protocol"
	"siguo/internal/room"
)

func TestStaticHandlerSetsCacheControlByAssetType(t *testing.T) {
	h := staticHandler()
	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/", want: "no-cache"},
		{path: "/app.js", want: "no-cache"},
		{path: "/app.css", want: "no-cache"},
		{path: "/song-small.mp3", want: "public, max-age=3600"},
		{path: "/setup-music.ogg", want: "public, max-age=3600"},
		{path: "/picture01.png", want: "public, max-age=3600"},
		{path: "/picture02.png", want: "public, max-age=3600"},
		{path: "/map3.jpg", want: "public, max-age=3600"},
		{path: "/map8.jpg", want: "public, max-age=3600"},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if got := rec.Header().Get("Cache-Control"); got != tc.want {
			t.Fatalf("%s cache-control = %q, want %q", tc.path, got, tc.want)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want %d", tc.path, rec.Code, http.StatusOK)
		}
	}
}

func TestCreateRoomStillAllowsLobbyWhenActiveRoomsAreFull(t *testing.T) {
	h := hub.New(1)
	active, err := h.Create()
	if err != nil {
		t.Fatalf("Create() active room error = %v", err)
	}
	startHTTPTestRoom(t, active)

	body, err := json.Marshal(protocol.CreateRoomRequest{Name: "late"})
	if err != nil {
		t.Fatalf("Marshal(CreateRoomRequest) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handleRooms(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestListActiveRooms(t *testing.T) {
	h := hub.New(25)
	active, err := h.Create()
	if err != nil {
		t.Fatalf("Create() active room error = %v", err)
	}
	startHTTPTestRoom(t, active)
	if _, err := h.Create(); err != nil {
		t.Fatalf("Create() lobby room error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	rec := httptest.NewRecorder()
	handleRooms(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response protocol.ActiveRoomsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal(active rooms) error = %v", err)
	}
	if len(response.Rooms) != 1 || response.Rooms[0].Code != active.Code {
		t.Fatalf("active rooms = %+v, want only %s", response.Rooms, active.Code)
	}
}

func TestListPopulatedLobbyRoomsForWatching(t *testing.T) {
	h := hub.New(25)
	lobby, err := h.Create()
	if err != nil {
		t.Fatalf("Create() lobby room error = %v", err)
	}
	if _, err := lobby.Join("north", ""); err != nil {
		t.Fatalf("Join(north) error = %v", err)
	}
	if _, err := lobby.Join("south", ""); err != nil {
		t.Fatalf("Join(south) error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	rec := httptest.NewRecorder()
	handleRooms(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var response protocol.ActiveRoomsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal(active rooms) error = %v", err)
	}
	if len(response.Rooms) != 1 || response.Rooms[0].Code != lobby.Code || response.Rooms[0].Phase != room.PhaseLobby {
		t.Fatalf("active rooms = %+v, want populated lobby %s", response.Rooms, lobby.Code)
	}
}

func startHTTPTestRoom(t *testing.T, r *room.Room) {
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
