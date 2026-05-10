package protocol

import "siguo/internal/game"

type TimeControl struct {
	SetupSeconds int `json:"setupSeconds"`
	MoveSeconds  int `json:"moveSeconds"`
	IncrementSec int `json:"incrementSec"`
}

func DefaultTimeControl() TimeControl {
	return TimeControl{SetupSeconds: 180, MoveSeconds: 45, IncrementSec: 0}
}

const MaxSkipsPerPlayer = 5

type PendingRequest struct {
	Kind  string      `json:"kind"`
	From  game.Seat   `json:"from"`
	Stage string      `json:"stage"`
	Acks  []game.Seat `json:"acks,omitempty"`
}

type CreateRoomRequest struct {
	Name          string        `json:"name"`
	Mode          game.GameMode `json:"mode,omitempty"`
	TimeControl   *TimeControl  `json:"timeControl,omitempty"`
	AllowTeamChat *bool         `json:"allowTeamChat,omitempty"`
}

type JoinRoomRequest struct {
	Name         string `json:"name"`
	SessionToken string `json:"sessionToken,omitempty"`
}

type JoinRoomResponse struct {
	Code         string    `json:"code"`
	SessionToken string    `json:"sessionToken"`
	Seat         game.Seat `json:"seat"`
	Name         string    `json:"name"`
	Host         bool      `json:"host"`
}

type ActiveRoomsResponse struct {
	Rooms []ActiveRoomInfo `json:"rooms"`
}

type ActiveRoomInfo struct {
	Code        string        `json:"code"`
	Mode        game.GameMode `json:"mode"`
	Phase       string        `json:"phase"`
	Seats       []SeatInfo    `json:"seats"`
	Viewers     int           `json:"viewers"`
	MaxViewers  int           `json:"maxViewers"`
	CanJoinView bool          `json:"canJoinView"`
}

type RoomSnapshot struct {
	Code            string             `json:"code"`
	Mode            game.GameMode      `json:"mode"`
	Phase           string             `json:"phase"`
	Seats           []SeatInfo         `json:"seats"`
	SelfSeat        game.Seat          `json:"selfSeat"`
	HostSeat        game.Seat          `json:"hostSeat"`
	AllowTeamChat   bool               `json:"allowTeamChat"`
	TimeControl     TimeControl        `json:"timeControl"`
	Turn            game.Seat          `json:"turn"`
	Winner          []game.Seat        `json:"winner,omitempty"`
	Ready           []game.Seat        `json:"ready,omitempty"`
	View            *game.ClientView   `json:"view,omitempty"`
	Skips           map[game.Seat]int  `json:"skips,omitempty"`
	MaxSkips        int                `json:"maxSkips,omitempty"`
	SetupDeadlineMs int64              `json:"setupDeadlineMs,omitempty"`
	MoveDeadlineMs  int64              `json:"moveDeadlineMs,omitempty"`
	MoveLimitSec    int                `json:"moveLimitSec,omitempty"`
	Eliminated      map[game.Seat]bool `json:"eliminated,omitempty"`
	Request         *PendingRequest    `json:"request,omitempty"`
}

type SeatInfo struct {
	Seat      game.Seat `json:"seat"`
	Name      string    `json:"name,omitempty"`
	Connected bool      `json:"connected"`
	Host      bool      `json:"host"`
	Ready     bool      `json:"ready"`
}

type ClientMessage struct {
	Type          string        `json:"type"`
	Seq           int64         `json:"seq,omitempty"`
	Name          string        `json:"name,omitempty"`
	SessionToken  string        `json:"sessionToken,omitempty"`
	TimeControl   *TimeControl  `json:"timeControl,omitempty"`
	AllowTeamChat *bool         `json:"allowTeamChat,omitempty"`
	Mode          game.GameMode `json:"mode,omitempty"`
	Seat          *game.Seat    `json:"seat,omitempty"`
	PieceID       game.PieceID  `json:"pieceId,omitempty"`
	Row           int           `json:"row,omitempty"`
	Col           int           `json:"col,omitempty"`
	From          game.Pos      `json:"from,omitempty"`
	To            game.Pos      `json:"to,omitempty"`
	Path          []game.Pos    `json:"path,omitempty"`
	Channel       string        `json:"channel,omitempty"`
	Text          string        `json:"text,omitempty"`
	Emote         string        `json:"emote,omitempty"`
	Kind          string        `json:"kind,omitempty"`
	Accept        bool          `json:"accept,omitempty"`
}

type ServerMessage struct {
	Type   string           `json:"type"`
	Seq    int64            `json:"seq,omitempty"`
	Room   *RoomSnapshot    `json:"room,omitempty"`
	View   *game.ClientView `json:"view,omitempty"`
	Event  *game.Event      `json:"event,omitempty"`
	Chat   *ChatMessage     `json:"chat,omitempty"`
	Error  *ErrorMessage    `json:"error,omitempty"`
	Notice string           `json:"notice,omitempty"`
}

type ChatMessage struct {
	From    game.Seat `json:"from"`
	Name    string    `json:"name"`
	Viewer  bool      `json:"viewer,omitempty"`
	Channel string    `json:"channel"`
	Text    string    `json:"text,omitempty"`
	Emote   string    `json:"emote,omitempty"`
	TS      int64     `json:"ts"`
}

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	RefSeq  int64  `json:"refSeq,omitempty"`
}
