package protocol

import "siguo/internal/game"

type TimeControl struct {
	SetupSeconds int `json:"setupSeconds"`
	MoveSeconds  int `json:"moveSeconds"`
	IncrementSec int `json:"incrementSec"`
}

func DefaultTimeControl() TimeControl {
	return TimeControl{SetupSeconds: 180, MoveSeconds: 30, IncrementSec: 5}
}

type CreateRoomRequest struct {
	Name          string       `json:"name"`
	TimeControl   *TimeControl `json:"timeControl,omitempty"`
	AllowTeamChat *bool        `json:"allowTeamChat,omitempty"`
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

type RoomSnapshot struct {
	Code          string           `json:"code"`
	Phase         string           `json:"phase"`
	Seats         []SeatInfo       `json:"seats"`
	HostSeat      game.Seat        `json:"hostSeat"`
	AllowTeamChat bool             `json:"allowTeamChat"`
	TimeControl   TimeControl      `json:"timeControl"`
	Turn          game.Seat        `json:"turn"`
	Winner        []game.Seat      `json:"winner,omitempty"`
	Ready         []game.Seat      `json:"ready,omitempty"`
	View          *game.ClientView `json:"view,omitempty"`
}

type SeatInfo struct {
	Seat      game.Seat `json:"seat"`
	Name      string    `json:"name,omitempty"`
	Connected bool      `json:"connected"`
	Host      bool      `json:"host"`
	Ready     bool      `json:"ready"`
}

type ClientMessage struct {
	Type          string       `json:"type"`
	Seq           int64        `json:"seq,omitempty"`
	Name          string       `json:"name,omitempty"`
	SessionToken  string       `json:"sessionToken,omitempty"`
	TimeControl   *TimeControl `json:"timeControl,omitempty"`
	AllowTeamChat *bool        `json:"allowTeamChat,omitempty"`
	PieceID       game.PieceID `json:"pieceId,omitempty"`
	Row           int          `json:"row,omitempty"`
	Col           int          `json:"col,omitempty"`
	From          game.Pos     `json:"from,omitempty"`
	To            game.Pos     `json:"to,omitempty"`
	Path          []game.Pos   `json:"path,omitempty"`
	Channel       string       `json:"channel,omitempty"`
	Text          string       `json:"text,omitempty"`
	Emote         string       `json:"emote,omitempty"`
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
