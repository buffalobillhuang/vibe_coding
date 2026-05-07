package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	mux.Handle("/ws", ws.Handler{Hub: h, AllowOrigin: cfg.AllowOrigin})
	mux.HandleFunc("/api/rooms", handleRooms(h))
	mux.HandleFunc("/api/rooms/", handleRoom(h))
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
	Addr        string
	MaxRooms    int
	LogLevel    string
	AllowOrigin string
}

func readConfig() config {
	cfg := config{
		Addr:        env("SIGUO_ADDR", ":8080"),
		MaxRooms:    envInt("SIGUO_MAX_ROOMS", 100),
		LogLevel:    env("SIGUO_LOG_LEVEL", "info"),
		AllowOrigin: env("SIGUO_ALLOW_ORIGIN", ""),
	}
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "listen address")
	flag.IntVar(&cfg.MaxRooms, "max-rooms", cfg.MaxRooms, "maximum active rooms")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "debug, info, warn, or error")
	flag.StringVar(&cfg.AllowOrigin, "allow-origin", cfg.AllowOrigin, "allowed websocket origin")
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
