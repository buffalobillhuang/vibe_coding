package ws

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"siguo/internal/hub"
)

func TestServeHTTP_LogsOpenAndCloseOnAbruptDisconnect(t *testing.T) {
	h := hub.New(0)
	room, err := h.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	player, err := room.Join("alice", "")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}

	var buf threadSafeBuffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := httptest.NewServer(Handler{Hub: h, Logger: logger})
	defer srv.Close()

	conn, br := dialWebSocket(t, srv.URL, room.Code, player.Token)
	// Read one frame so we know the server's writer goroutine is alive.
	rw := bufio.NewReadWriter(br, bufio.NewWriter(conn))
	if _, _, err := readFrame(rw); err != nil {
		t.Fatalf("read first frame: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close conn: %v", err)
	}

	waitForLog(t, &buf, "ws.close", 2*time.Second)
	logs := buf.String()
	for _, want := range []string{
		"msg=ws.open",
		"msg=ws.close",
		"room=" + room.Code,
		"role=player",
		"reason=",
		"age_ms=",
	} {
		if !strings.Contains(logs, want) {
			t.Fatalf("missing %q in logs:\n%s", want, logs)
		}
	}
}

func TestServeHTTP_LogsRejectOnMissingToken(t *testing.T) {
	h := hub.New(0)
	room, err := h.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	var buf threadSafeBuffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := httptest.NewServer(Handler{Hub: h, Logger: logger})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "?room=" + room.Code)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()

	if got := buf.String(); !strings.Contains(got, "reason=missing_token") {
		t.Fatalf("expected reason=missing_token; got:\n%s", got)
	}
}

func TestServeHTTP_LogsRejectOnUnknownRoom(t *testing.T) {
	h := hub.New(0)
	var buf threadSafeBuffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	srv := httptest.NewServer(Handler{Hub: h, Logger: logger})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "?room=NOPE99&token=anything")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()

	if got := buf.String(); !strings.Contains(got, "reason=room_not_found") {
		t.Fatalf("expected reason=room_not_found; got:\n%s", got)
	}
}

func dialWebSocket(t *testing.T, baseURL, code, token string) (net.Conn, *bufio.Reader) {
	t.Helper()
	u, _ := url.Parse(baseURL)
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		t.Fatalf("rand: %v", err)
	}
	key := base64.StdEncoding.EncodeToString(keyBytes)
	req := "GET /?room=" + code + "&token=" + token + " HTTP/1.1\r\n" +
		"Host: " + u.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write upgrade: %v", err)
	}
	br := bufio.NewReader(conn)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			t.Fatalf("read upgrade response: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}
	return conn, br
}

func waitForLog(t *testing.T, buf *threadSafeBuffer, needle string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(buf.String(), needle) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q in logs:\n%s", needle, buf.String())
}

type threadSafeBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (t *threadSafeBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.b.Write(p)
}

func (t *threadSafeBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.b.String()
}
