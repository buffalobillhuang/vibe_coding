package ws

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
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
)

type Handler struct {
	Hub         *hub.Hub
	AllowOrigin string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Hub == nil {
		http.Error(w, "hub not configured", http.StatusInternalServerError)
		return
	}
	if !h.originAllowed(r) {
		http.Error(w, "origin forbidden", http.StatusForbidden)
		return
	}
	code := r.URL.Query().Get("room")
	token := r.URL.Query().Get("token")
	room, ok := h.Hub.Get(code)
	if !ok || token == "" {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	out, _, err := room.Connect(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, rw, err := accept(w, r)
	if err != nil {
		room.Disconnect(token)
		return
	}
	defer conn.Close()
	defer room.Disconnect(token)

	done := make(chan struct{})
	var writeMu sync.Mutex
	safeWrite := func(op byte, payload []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return writeFrame(rw, op, payload)
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case msg, ok := <-out:
				if !ok {
					_ = safeWrite(opClose, nil)
					return
				}
				if err := safeWrite(opText, msg); err != nil {
					return
				}
			case <-ticker.C:
				if err := safeWrite(opPing, []byte("ping")); err != nil {
					return
				}
			}
		}
	}()

	for {
		op, payload, err := readFrame(rw)
		if err != nil {
			return
		}
		switch op {
		case opText:
			room.Handle(token, payload)
		case opPing:
			_ = safeWrite(opPong, payload)
		case opClose:
			return
		}
		select {
		case <-done:
			return
		default:
		}
	}
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
