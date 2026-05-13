package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestOriginTrackerDedupesAndSorts(t *testing.T) {
	tr := newOriginTracker()
	tr.add("https://b.example")
	tr.add("http://a.example")
	tr.add("https://b.example") // duplicate
	tr.add("")                  // empty
	tr.add("   ")               // whitespace
	tr.add("http://c.example/") // trailing slash

	got := tr.list()
	want := []string{"http://a.example", "http://c.example", "https://b.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list = %v, want %v", got, want)
	}
}

func TestHandleOriginsReturnsTrackedList(t *testing.T) {
	tr := newOriginTracker()
	tr.add("http://1.2.3.4")
	tr.add("https://example.test")

	req := httptest.NewRequest(http.MethodGet, "/api/origins", nil)
	rec := httptest.NewRecorder()
	handleOrigins(tr).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp struct {
		Origins []string `json:"origins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	want := []string{"http://1.2.3.4", "https://example.test"}
	if !reflect.DeepEqual(resp.Origins, want) {
		t.Fatalf("origins = %v, want %v", resp.Origins, want)
	}
}

func TestDetectPublicIPReturnsFirstValidResponse(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not-an-ip\n"))
	}))
	defer bad.Close()
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("203.0.113.7\n"))
	}))
	defer good.Close()

	got := detectPublicIP(context.Background(), []string{bad.URL, good.URL})
	if got != "203.0.113.7" {
		t.Fatalf("ip = %q, want 203.0.113.7", got)
	}
}

func TestEnvBoolParsesCommonTruthyValues(t *testing.T) {
	for _, v := range []string{"true", "TRUE", "1", "yes", "on", "On"} {
		t.Setenv("SIGUO_TEST_FLAG", v)
		if !envBool("SIGUO_TEST_FLAG", false) {
			t.Fatalf("envBool(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"false", "0", "no", "off", "FALSE"} {
		t.Setenv("SIGUO_TEST_FLAG", v)
		if envBool("SIGUO_TEST_FLAG", true) {
			t.Fatalf("envBool(%q) = true, want false", v)
		}
	}
	t.Setenv("SIGUO_TEST_FLAG", "")
	if !envBool("SIGUO_TEST_FLAG", true) {
		t.Fatal("envBool(empty) should fall back to default true")
	}
	if envBool("SIGUO_TEST_FLAG", false) {
		t.Fatal("envBool(empty) should fall back to default false")
	}
}

func TestDetectPublicIPReturnsEmptyWhenAllFail(t *testing.T) {
	// Closed server; both attempts should fail at the dial step.
	bad1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("garbage"))
	}))
	bad1.Close()
	bad2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(""))
	}))
	bad2.Close()

	got := detectPublicIP(context.Background(), []string{bad1.URL, bad2.URL})
	if got != "" {
		t.Fatalf("ip = %q, want empty", got)
	}
}


func TestDetectLANOriginsPrefersOutboundPrivateIPv4(t *testing.T) {
	got := detectLANOrigins(":8080", []lanInterfaceAddr{
		{Name: "bridge100", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.64.1")},
		{Name: "en0", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.1.55")},
	}, net.ParseIP("192.168.1.55"), false)
	want := []string{"http://192.168.1.55:8080"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("origins = %v, want %v", got, want)
	}
}

func TestDetectLANOriginsUsesExplicitListenerHost(t *testing.T) {
	got := detectLANOrigins("192.168.1.77:1080", nil, nil, false)
	want := []string{"http://192.168.1.77:1080"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("origins = %v, want %v", got, want)
	}
}


func TestDetectLANOriginsRanksPhysicalInterfaceAboveBridge(t *testing.T) {
	got := detectLANOrigins(":8080", []lanInterfaceAddr{
		{Name: "bridge100", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.64.1")},
		{Name: "en0", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.3.157")},
	}, nil, false)
	want := []string{"http://192.168.3.157:8080"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("origins = %v, want %v", got, want)
	}
}


func TestDetectLANOriginsFallsBackToSinglePrivateIPv4(t *testing.T) {
	got := detectLANOrigins(":8080", []lanInterfaceAddr{
		{Name: "en0", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.1.55")},
	}, nil, false)
	want := []string{"http://192.168.1.55:8080"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("origins = %v, want %v", got, want)
	}
}

func TestDetectLANOriginsSkipsLoopbackAndContainers(t *testing.T) {
	privateAddrs := []lanInterfaceAddr{
		{Name: "en0", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.1.55")},
	}
	if got := detectLANOrigins("127.0.0.1:8080", privateAddrs, nil, false); len(got) != 0 {
		t.Fatalf("loopback listener origins = %v, want empty", got)
	}
	if got := detectLANOrigins(":8080", privateAddrs, nil, true); len(got) != 0 {
		t.Fatalf("container origins = %v, want empty", got)
	}
}

func TestDetectLANOriginsSkipsOnlyVirtualPrivateInterfaces(t *testing.T) {
	got := detectLANOrigins(":8080", []lanInterfaceAddr{
		{Name: "bridge100", Flags: net.FlagUp | net.FlagBroadcast, IP: net.ParseIP("192.168.64.1")},
		{Name: "utun4", Flags: net.FlagUp | net.FlagPointToPoint, IP: net.ParseIP("10.0.0.8")},
	}, nil, false)
	if len(got) != 0 {
		t.Fatalf("origins = %v, want empty", got)
	}
}
