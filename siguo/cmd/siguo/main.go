package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"siguo/internal/hub"
	"siguo/internal/protocol"
	"siguo/internal/static"
	"siguo/internal/ws"
)

func main() {
	cfg := readConfig()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.logLevel()}))
	h := hub.New(cfg.MaxRooms)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ready\n"))
	})
	mux.Handle("/ws", ws.Handler{Hub: h, AllowOrigin: cfg.AllowOrigin, Logger: logger})
	mux.HandleFunc("/api/rooms", handleRooms(h))
	mux.HandleFunc("/api/rooms/", handleRoom(h))

	tracker := newOriginTracker()
	for _, o := range strings.Split(os.Getenv("SIGUO_ALTERNATE_ORIGINS"), ",") {
		tracker.add(o)
	}
	for _, o := range detectLANOrigins(cfg.Addr, localLANInterfaceAddrs(), detectPreferredOutboundIP(), runningInContainer()) {
		tracker.add(o)
	}
	if cfg.DetectPublicIP {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if ip := detectPublicIP(ctx, publicIPServices); ip != "" {
				tracker.add("http://" + ip)
				logger.Info("public IP detected for invite-link alternates", "ip", ip)
			} else {
				logger.Info("public IP detection failed; invite links will use only the page origin")
			}
		}()
	}
	mux.HandleFunc("/api/origins", handleOrigins(tracker))

	mux.Handle("/", staticHandler())

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           logMiddleware(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("siguo listening", "addr", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

type config struct {
	Addr            string
	MaxRooms        int
	LogLevel        string
	AllowOrigin     string
	DetectPublicIP  bool
}

func readConfig() config {
	cfg := config{
		Addr:           env("SIGUO_ADDR", ":8080"),
		MaxRooms:       envInt("SIGUO_MAX_ROOMS", 25),
		LogLevel:       env("SIGUO_LOG_LEVEL", "info"),
		AllowOrigin:    env("SIGUO_ALLOW_ORIGIN", ""),
		DetectPublicIP: envBool("SIGUO_DETECT_PUBLIC_IP", false),
	}
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	flag.IntVar(&cfg.MaxRooms, "max-rooms", cfg.MaxRooms, "maximum active rooms")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "debug, info, warn, or error")
	flag.StringVar(&cfg.AllowOrigin, "allow-origin", cfg.AllowOrigin, "allowed websocket origin")
	flag.BoolVar(&cfg.DetectPublicIP, "detect-public-ip", cfg.DetectPublicIP, "advertise the server's public egress IP as an invite-link alternate (cloud-with-Caddy deploys; off for LAN)")
	flag.Bool("persist", false, "reserved for future SQLite persistence")
	flag.String("db-path", "./siguo.db", "reserved SQLite path")
	flag.Bool("metrics", false, "reserved for Prometheus metrics")
	flag.Parse()
	return cfg
}

func (c config) logLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func handleRooms(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			writeJSON(w, protocol.ActiveRoomsResponse{Rooms: h.ActiveRooms()})
			return
		}
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var req protocol.CreateRoomRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		room, err := h.Create()
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		room.ConfigureInitial(req.Mode, req.TimeControl, req.AllowTeamChat)
		player, err := room.Join(req.Name, "")
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, protocol.JoinRoomResponse{
			Code:         room.Code,
			SessionToken: player.Token,
			Seat:         player.Seat,
			Name:         player.Name,
			Host:         player.Host,
		})
	}
}

func handleRoom(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/rooms/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.NotFound(w, r)
			return
		}
		room, ok := h.Get(parts[0])
		if !ok {
			writeError(w, http.StatusNotFound, "room not found")
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			token := r.URL.Query().Get("token")
			snap, err := room.SnapshotFor(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, err.Error())
				return
			}
			writeJSON(w, snap)
			return
		}
		if len(parts) == 2 && parts[1] == "join" && r.Method == http.MethodPost {
			var req protocol.JoinRoomRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			player, err := room.Join(req.Name, req.SessionToken)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeJSON(w, protocol.JoinRoomResponse{
				Code:         room.Code,
				SessionToken: player.Token,
				Seat:         player.Seat,
				Name:         player.Name,
				Host:         player.Host,
			})
			return
		}
		http.NotFound(w, r)
	}
}

func staticHandler() http.Handler {
	sub, err := fs.Sub(static.FS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", staticCacheControl(r.URL.Path))
		if r.URL.Path != "/" {
			if f, err := sub.Open(strings.TrimPrefix(r.URL.Path, "/")); err == nil {
				_ = f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func staticCacheControl(path string) string {
	switch {
	case path == "/", strings.HasSuffix(path, ".html"), strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".css"):
		return "no-cache"
	case strings.HasSuffix(path, ".mp3"), strings.HasSuffix(path, ".ogg"), strings.HasSuffix(path, ".png"), strings.HasSuffix(path, ".jpg"):
		return "public, max-age=3600"
	default:
		return "no-cache"
	}
}

func logMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Debug("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
		fmt.Fprintf(os.Stderr, "invalid %s=%q, using %d\n", key, v, fallback)
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		fmt.Fprintf(os.Stderr, "invalid %s=%q, using %v\n", key, os.Getenv(key), fallback)
		return fallback
	}
}

type lanInterfaceAddr struct {
	Name  string
	Flags net.Flags
	IP    net.IP
}

func localLANInterfaceAddrs() []lanInterfaceAddr {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []lanInterfaceAddr
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := addrIP(addr)
			if ip == nil || ip.To4() == nil {
				continue
			}
			out = append(out, lanInterfaceAddr{Name: iface.Name, Flags: iface.Flags, IP: ip})
		}
	}
	return out
}

func runningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return strings.TrimSpace(os.Getenv("container")) != ""
}

func detectPreferredOutboundIP() net.IP {
	for _, target := range []string{"1.1.1.1:80", "8.8.8.8:80", "9.9.9.9:80"} {
		conn, err := net.Dial("udp", target)
		if err != nil {
			continue
		}
		addr, _ := conn.LocalAddr().(*net.UDPAddr)
		_ = conn.Close()
		if addr == nil || addr.IP == nil || addr.IP.To4() == nil || !addr.IP.IsPrivate() {
			continue
		}
		return addr.IP
	}
	return nil
}

func detectLANOrigins(listenAddr string, addrs []lanInterfaceAddr, preferredIP net.IP, inContainer bool) []string {
	if inContainer {
		return nil
	}
	host, port := splitListenAddr(listenAddr)
	if isLoopbackHost(host) {
		return nil
	}
	if isExplicitHost(host) {
		return []string{originForHostPort(host, port)}
	}
	if preferredIP != nil && preferredIP.To4() != nil && preferredIP.IsPrivate() {
		if origin := originForHostPort(preferredIP.String(), port); origin != "" {
			return []string{origin}
		}
	}
	bestOrigin := ""
	bestScore := 0
	tied := false
	for _, addr := range addrs {
		if !isLikelyLANInterface(addr) {
			continue
		}
		origin := originForHostPort(addr.IP.String(), port)
		if origin == "" {
			continue
		}
		score := lanInterfaceScore(addr)
		if score <= 0 {
			continue
		}
		if bestOrigin == "" || score > bestScore {
			bestOrigin = origin
			bestScore = score
			tied = false
			continue
		}
		if score == bestScore && origin != bestOrigin {
			tied = true
		}
	}
	if bestOrigin != "" && !tied {
		return []string{bestOrigin}
	}
	return nil
}

func isLikelyLANInterface(addr lanInterfaceAddr) bool {
	if addr.IP == nil || addr.IP.IsLoopback() || addr.IP.To4() == nil || !addr.IP.IsPrivate() {
		return false
	}
	if addr.Flags&net.FlagUp == 0 || addr.Flags&net.FlagLoopback != 0 || addr.Flags&net.FlagPointToPoint != 0 {
		return false
	}
	return true
}

func lanInterfaceScore(addr lanInterfaceAddr) int {
	name := strings.ToLower(strings.TrimSpace(addr.Name))
	score := 0
	if addr.Flags&net.FlagBroadcast != 0 {
		score += 10
	}
	switch {
	case strings.HasPrefix(name, "en"):
		score += 100
	case strings.HasPrefix(name, "eth"), strings.HasPrefix(name, "wlan"), strings.HasPrefix(name, "wifi"), strings.HasPrefix(name, "wl"):
		score += 90
	case strings.HasPrefix(name, "bridge"), strings.HasPrefix(name, "docker"), strings.HasPrefix(name, "br-"), strings.HasPrefix(name, "virbr"), strings.HasPrefix(name, "vmnet"), strings.HasPrefix(name, "vboxnet"), strings.HasPrefix(name, "cni"), strings.HasPrefix(name, "utun"), strings.HasPrefix(name, "tun"), strings.HasPrefix(name, "tap"), strings.HasPrefix(name, "zt"), strings.HasPrefix(name, "tailscale"), strings.HasPrefix(name, "awdl"), strings.HasPrefix(name, "llw"), strings.HasPrefix(name, "anpi"), strings.HasPrefix(name, "gif"), strings.HasPrefix(name, "stf"):
		score -= 100
	}
	return score
}

func splitListenAddr(listenAddr string) (string, string) {
	listenAddr = strings.TrimSpace(listenAddr)
	if listenAddr == "" {
		return "", ""
	}
	host, port, err := net.SplitHostPort(listenAddr)
	if err == nil {
		return strings.TrimSpace(host), strings.TrimSpace(port)
	}
	if strings.HasPrefix(listenAddr, ":") {
		return "", strings.TrimPrefix(listenAddr, ":")
	}
	return listenAddr, ""
}

func isExplicitHost(host string) bool {
	host = normalizedHost(host)
	return host != "" && host != "0.0.0.0" && host != "::"
}

func isLoopbackHost(host string) bool {
	host = normalizedHost(host)
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func normalizedHost(host string) string {
	host = strings.TrimSpace(host)
	return strings.Trim(host, "[]")
}

func addrIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPAddr:
		return v.IP
	case *net.IPNet:
		return v.IP
	default:
		return nil
	}
}

func originForHostPort(host, port string) string {
	host = normalizedHost(host)
	if host == "" {
		return ""
	}
	if port == "" || port == "80" {
		if strings.Contains(host, ":") {
			return "http://[" + host + "]"
		}
		return "http://" + host
	}
	return "http://" + net.JoinHostPort(host, port)
}

// originTracker collects origin URLs (scheme + host) that the server should
// advertise as alternate share endpoints for invite links. Pre-seeded from
// SIGUO_ALTERNATE_ORIGINS and augmented at startup with the server's detected
// public IP.
type originTracker struct {
	mu      sync.Mutex
	origins map[string]struct{}
}

func newOriginTracker() *originTracker {
	return &originTracker{origins: map[string]struct{}{}}
}

func (t *originTracker) add(origin string) {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.origins[origin] = struct{}{}
}

func (t *originTracker) list() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, 0, len(t.origins))
	for o := range t.origins {
		out = append(out, o)
	}
	sort.Strings(out)
	return out
}

func handleOrigins(t *originTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string][]string{"origins": t.list()})
	}
}

// publicIPServices is the ordered fallback list for outbound IP detection.
// Each service must return the egress IPv4 as a plain text body.
var publicIPServices = []string{
	"https://api.ipify.org",
	"https://ifconfig.me/ip",
	"https://checkip.amazonaws.com",
}

// detectPublicIP tries each service in order and returns the first response
// that parses as a valid IP. Returns "" if everything fails — callers should
// treat that as "no alternate available" and degrade gracefully.
func detectPublicIP(ctx context.Context, services []string) string {
	client := &http.Client{Timeout: 4 * time.Second}
	for _, url := range services {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
		resp.Body.Close()
		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}
