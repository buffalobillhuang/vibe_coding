package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"siguo/internal/hub"
)

const (
	opText  = 1
	opClose = 8
	opPing  = 9
	opPong  = 10

	pingInterval = 3 * time.Second
	pongWait     = 12 * time.Second
	writeWait    = 10 * time.Second
)

var heartbeatPayload = []byte(`{"type":"heartbeat"}`)

const rejectThrottleWindow = time.Minute

var rejectThrottle = struct {
	mu   sync.Mutex
	last map[string]time.Time
}{last: map[string]time.Time{}}

func shouldLogRejectInfo(code string) bool {
	rejectThrottle.mu.Lock()
	defer rejectThrottle.mu.Unlock()
	now := time.Now()
	for k, t := range rejectThrottle.last {
		if now.Sub(t) > rejectThrottleWindow*5 {
			delete(rejectThrottle.last, k)
		}
	}
	if last, ok := rejectThrottle.last[code]; ok && now.Sub(last) < rejectThrottleWindow {
		return false
	}
	rejectThrottle.last[code] = now
	return true
}

type Handler struct {
	Hub         *hub.Hub
	AllowOrigin string
	Logger      *slog.Logger
}

func (h Handler) logger() *slog.Logger {
	if h.Logger != nil {
		return h.Logger
	}
	return slog.Default()
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.logger()
	if h.Hub == nil {
		http.Error(w, "hub not configured", http.StatusInternalServerError)
		return
	}
	if !h.originAllowed(r) {
		logger.Warn("ws.reject", "reason", "origin_forbidden", "origin", r.Header.Get("Origin"), "remote", r.RemoteAddr)
		http.Error(w, "origin forbidden", http.StatusForbidden)
		return
	}
	code := r.URL.Query().Get("room")
	token := r.URL.Query().Get("token")
	room, ok := h.Hub.Get(code)
	if !ok {
		if shouldLogRejectInfo(code) {
			logger.Info("ws.reject", "reason", "room_not_found", "room", code, "remote", r.RemoteAddr)
		} else {
			logger.Debug("ws.reject", "reason", "room_not_found", "room", code, "remote", r.RemoteAddr, "throttled", true)
		}
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	isViewer := r.URL.Query().Get("viewer") == "1"
	viewerName := r.URL.Query().Get("name")
	role := "player"
	if isViewer {
		role = "viewer"
	}
	var out chan []byte
	var viewerID string
	var err error
	if isViewer {
		out, viewerID, err = room.ConnectViewer(viewerName)
	} else {
		if token == "" {
			logger.Info("ws.reject", "reason", "missing_token", "room", code, "remote", r.RemoteAddr)
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		out, _, err = room.Connect(token)
	}
	if err != nil {
		logger.Warn("ws.reject", "reason", "connect_failed", "role", role, "room", code, "remote", r.RemoteAddr, "error", err.Error())
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, rw, err := accept(w, r)
	if err != nil {
		logger.Warn("ws.reject", "reason", "handshake_failed", "role", role, "room", code, "remote", r.RemoteAddr, "error", err.Error())
		if isViewer {
			room.DisconnectViewer(viewerID)
		} else {
			room.Disconnect(token, out)
		}
		return
	}
	defer conn.Close()
	if isViewer {
		viewerName = room.ViewerName(viewerID)
		room.BroadcastViewerNotice(viewerName, "加入观战")
	}
	connID := connIDOf(token, viewerID)
	connStart := time.Now()
	logger.Info("ws.open", "role", role, "room", code, "conn", connID, "remote", r.RemoteAddr)
	closeReason := "client_close"
	if isViewer {
		defer func() {
			logger.Info("ws.close", "role", role, "room", code, "conn", connID, "reason", closeReason, "age_ms", time.Since(connStart).Milliseconds())
			room.BroadcastViewerNotice(viewerName, "离开观战")
			room.DisconnectViewer(viewerID)
		}()
	} else {
		defer func() {
			logger.Info("ws.close", "role", role, "room", code, "conn", connID, "reason", closeReason, "age_ms", time.Since(connStart).Milliseconds())
			room.Disconnect(token, out)
		}()
	}
	refreshReadDeadline := func() {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	}

	done := make(chan struct{})
	var writeMu sync.Mutex
	safeWrite := func(op byte, payload []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return writeFrame(rw, op, payload)
	}
	var writerReason string
	refreshReadDeadline()
	go func() {
		defer close(done)
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case msg, ok := <-out:
				if !ok {
					writerReason = "server_replaced"
					_ = safeWrite(opClose, nil)
					_ = conn.Close()
					return
				}
				if err := safeWrite(opText, msg); err != nil {
					writerReason = "write_err:" + err.Error()
					logger.Info("ws.write_err", "role", role, "room", code, "conn", connID, "stage", "text", "error", err.Error(), "age_ms", time.Since(connStart).Milliseconds())
					_ = conn.Close()
					return
				}
			case <-ticker.C:
				if err := safeWrite(opPing, []byte("ping")); err != nil {
					writerReason = "ping_err:" + err.Error()
					logger.Info("ws.write_err", "role", role, "room", code, "conn", connID, "stage", "ping", "error", err.Error(), "age_ms", time.Since(connStart).Milliseconds())
					_ = conn.Close()
					return
				}
				if err := safeWrite(opText, heartbeatPayload); err != nil {
					writerReason = "heartbeat_err:" + err.Error()
					logger.Info("ws.write_err", "role", role, "room", code, "conn", connID, "stage", "heartbeat", "error", err.Error(), "age_ms", time.Since(connStart).Milliseconds())
					_ = conn.Close()
					return
				}
			}
		}
	}()

	for {
		op, payload, err := readFrame(rw)
		if err != nil {
			if writerReason != "" {
				closeReason = writerReason
			} else {
				closeReason = "read_err:" + err.Error()
				logger.Info("ws.read_err", "role", role, "room", code, "conn", connID, "error", err.Error(), "age_ms", time.Since(connStart).Milliseconds())
			}
			return
		}
		refreshReadDeadline()
		switch op {
		case opText:
			if isViewer {
				room.HandleViewer(viewerID, payload)
			} else {
				room.Handle(token, payload)
			}
		case opPing:
			_ = safeWrite(opPong, payload)
		case opClose:
			closeReason = "client_close"
			return
		}
		select {
		case <-done:
			if writerReason != "" {
				closeReason = writerReason
			}
			return
		default:
		}
	}
}

func connIDOf(token, viewerID string) string {
	src := token
	if src == "" {
		src = viewerID
	}
	if len(src) > 8 {
		return src[:8]
	}
	return src
}

func (h Handler) originAllowed(r *http.Request) bool {
	if h.AllowOrigin == "" {
		return true
	}
	origin := r.Header.Get("Origin")
	return origin == "" || origin == h.AllowOrigin
}

func accept(w http.ResponseWriter, r *http.Request) (net.Conn, *bufio.ReadWriter, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "upgrade required", http.StatusUpgradeRequired)
		return nil, nil, errors.New("upgrade required")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing websocket key", http.StatusBadRequest)
		return nil, nil, errors.New("missing websocket key")
	}
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack unsupported", http.StatusInternalServerError)
		return nil, nil, errors.New("hijack unsupported")
	}
	conn, rw, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	acceptKey := websocketAccept(key)
	_, err = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n")
	if err == nil {
		err = rw.Flush()
	}
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}
	return conn, rw, nil
}

func websocketAccept(key string) string {
	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func readFrame(rw *bufio.ReadWriter) (byte, []byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(rw, header); err != nil {
		return 0, nil, err
	}
	op := header[0] & 0x0f
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7f)
	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(rw, ext[:]); err != nil {
			return 0, nil, err
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(rw, ext[:]); err != nil {
			return 0, nil, err
		}
		length = binary.BigEndian.Uint64(ext[:])
	}
	if length > 64*1024 {
		return 0, nil, errors.New("frame too large")
	}
	var mask [4]byte
	if masked {
		if _, err := io.ReadFull(rw, mask[:]); err != nil {
			return 0, nil, err
		}
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(rw, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return op, payload, nil
}

func writeFrame(rw *bufio.ReadWriter, op byte, payload []byte) error {
	header := []byte{0x80 | op}
	length := len(payload)
	switch {
	case length < 126:
		header = append(header, byte(length))
	case length <= 0xffff:
		header = append(header, 126, 0, 0)
		binary.BigEndian.PutUint16(header[len(header)-2:], uint16(length))
	default:
		header = append(header, 127, 0, 0, 0, 0, 0, 0, 0, 0)
		binary.BigEndian.PutUint64(header[len(header)-8:], uint64(length))
	}
	if _, err := rw.Write(header); err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := rw.Write(payload); err != nil {
			return err
		}
	}
	return rw.Flush()
}
