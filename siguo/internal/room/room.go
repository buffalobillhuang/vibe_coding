package room

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	mrand "math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"siguo/internal/game"
	"siguo/internal/protocol"
)

const (
	PhaseLobby   = "lobby"
	PhaseSetup   = "setup"
	PhasePlaying = "playing"
	PhaseEnded   = "ended"

	ChannelAll  = "all"
	ChannelTeam = "team"
)

var ErrRoomFull = errors.New("room is full")

type Player struct {
	Name      string
	Token     string
	Seat      game.Seat
	Host      bool
	Connected bool
	Ready     bool
	PieceIDs  []game.PieceID
	LastSeq   int64
}

type Room struct {
	mu            sync.Mutex
	Code          string
	phase         string
	mode          game.GameMode
	host          game.Seat
	allowTeamChat bool
	timeControl   protocol.TimeControl
	players       map[string]*Player
	seats         map[game.Seat]*Player
	connections   map[string]chan []byte
	pieces        map[game.PieceID]game.Piece
	placements    map[game.PieceID]game.Pos
	state         *game.GameState
	nextPieceID   game.PieceID
	serverSeq     int64
	chatWindow    map[string][]time.Time
	skips         map[game.Seat]int
	turnTimer     *time.Timer
	turnDeadline  time.Time
	turnEpoch     int64
	request       *protocol.PendingRequest
}

func New(code string) *Room {
	return &Room{
		Code:          code,
		phase:         PhaseLobby,
		mode:          game.ModeSiguo,
		allowTeamChat: true,
		timeControl:   protocol.DefaultTimeControl(),
		players:       map[string]*Player{},
		seats:         map[game.Seat]*Player{},
		connections:   map[string]chan []byte{},
		pieces:        map[game.PieceID]game.Piece{},
		placements:    map[game.PieceID]game.Pos{},
		nextPieceID:   1,
		chatWindow:    map[string][]time.Time{},
		skips:         map[game.Seat]int{},
	}
}

func (r *Room) Join(name, token string) (*Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = cleanName(name)
	if token != "" {
		if player := r.players[token]; player != nil {
			player.Name = name
			return player, nil
		}
	}
	if r.phase != PhaseLobby {
		return nil, errors.New("game already started")
	}
	if len(r.seats) >= len(game.ActiveSeats(r.mode)) {
		return nil, ErrRoomFull
	}

	seat := firstOpenSeat(r.mode, r.seats)
	if token == "" {
		token = newToken()
	}
	player := &Player{
		Name:  name,
		Token: token,
		Seat:  seat,
		Host:  len(r.seats) == 0,
	}
	if player.Host {
		r.host = seat
	}
	r.players[token] = player
	r.seats[seat] = player
	r.broadcastRoomLocked()
	return player, nil
}

func (r *Room) ConfigureInitial(mode game.GameMode, tc *protocol.TimeControl, allowTeamChat *bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseLobby {
		return
	}
	if mode == game.ModeJunqi {
		r.mode = game.ModeJunqi
		r.allowTeamChat = false
	} else {
		r.mode = game.ModeSiguo
	}
	if tc != nil {
		r.timeControl = *tc
	}
	if allowTeamChat != nil {
		r.allowTeamChat = *allowTeamChat
	}
	if r.mode == game.ModeJunqi {
		r.allowTeamChat = false
	}
}

func (r *Room) Connect(token string) (<-chan []byte, *Player, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.players[token]
	if player == nil {
		return nil, nil, errors.New("invalid session token")
	}
	ch := make(chan []byte, 32)
	r.connections[token] = ch
	player.Connected = true
	r.sendLocked(token, protocol.ServerMessage{Type: "room.state", Room: r.snapshotLocked(player.Seat)})
	r.sendLocked(token, protocol.ServerMessage{Type: "view", View: r.viewForLocked(player.Seat)})
	r.broadcastRoomLocked()
	return ch, player, nil
}

func (r *Room) Disconnect(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if ch := r.connections[token]; ch != nil {
		close(ch)
		delete(r.connections, token)
	}
	if player := r.players[token]; player != nil {
		player.Connected = false
	}
	r.broadcastRoomLocked()
}

func (r *Room) Handle(token string, raw []byte) {
	var msg protocol.ClientMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		r.SendError(token, "bad_json", "消息格式错误", 0)
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.players[token]
	if player == nil {
		r.sendErrorLocked(token, "unauthorized", "会话已失效", msg.Seq)
		return
	}
	if msg.Seq != 0 && msg.Seq <= player.LastSeq {
		r.sendLocked(token, protocol.ServerMessage{Type: "room.state", Room: r.snapshotLocked(player.Seat)})
		return
	}
	if msg.Seq != 0 {
		player.LastSeq = msg.Seq
	}

	switch msg.Type {
	case "room.config":
		r.handleConfigLocked(player, msg)
	case "room.start":
		r.handleStartLocked(player, msg.Seq)
	case "setup.randomize":
		r.handleRandomizeLocked(player, msg.Seq)
	case "setup.place":
		r.handlePlaceLocked(player, msg)
	case "setup.submit":
		r.handleSubmitLocked(player, msg.Seq)
	case "move":
		r.handleMoveLocked(player, msg)
	case "move.skip":
		r.handleSkipLocked(player, msg.Seq)
	case "request.tie":
		r.handleRequestLocked(player, "tie", msg.Seq)
	case "request.surrender":
		r.handleRequestLocked(player, "surrender", msg.Seq)
	case "request.respond":
		r.handleRequestRespondLocked(player, msg)
	case "request.cancel":
		r.handleRequestCancelLocked(player, msg.Seq)
	case "concede":
		r.handleConcedeLocked(player)
	case "chat.send":
		r.handleChatLocked(player, msg)
	default:
		r.sendErrorLocked(token, "unknown_type", "未知消息类型", msg.Seq)
	}
}

func (r *Room) SendError(token, code, message string, refSeq int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sendErrorLocked(token, code, message, refSeq)
}

func (r *Room) SnapshotFor(token string) (*protocol.RoomSnapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	player := r.players[token]
	if player == nil {
		return nil, errors.New("invalid session token")
	}
	return r.snapshotLocked(player.Seat), nil
}

func (r *Room) handleConfigLocked(player *Player, msg protocol.ClientMessage) {
	if !player.Host || r.phase != PhaseLobby {
		r.sendErrorLocked(player.Token, "forbidden", "只有房主可以在大厅修改设置", msg.Seq)
		return
	}
	if msg.Mode == game.ModeJunqi && len(r.seats) <= 1 {
		r.mode = game.ModeJunqi
		r.allowTeamChat = false
	} else if msg.Mode == game.ModeSiguo && len(r.seats) <= 1 {
		r.mode = game.ModeSiguo
	}
	if msg.TimeControl != nil {
		r.timeControl = *msg.TimeControl
	}
	if msg.AllowTeamChat != nil {
		r.allowTeamChat = *msg.AllowTeamChat
	}
	if r.mode == game.ModeJunqi {
		r.allowTeamChat = false
	}
	r.broadcastRoomLocked()
}

func (r *Room) handleStartLocked(player *Player, refSeq int64) {
	if !player.Host {
		r.sendErrorLocked(player.Token, "forbidden", "只有房主可以开始", refSeq)
		return
	}
	if len(r.seats) != len(game.ActiveSeats(r.mode)) {
		if r.mode == game.ModeJunqi {
			r.sendErrorLocked(player.Token, "not_ready", "需要两名玩家", refSeq)
		} else {
			r.sendErrorLocked(player.Token, "not_ready", "需要四名玩家", refSeq)
		}
		return
	}
	if r.phase != PhaseLobby {
		return
	}

	r.phase = PhaseSetup
	r.pieces = map[game.PieceID]game.Piece{}
	r.placements = map[game.PieceID]game.Pos{}
	for _, seat := range game.ActiveSeats(r.mode) {
		p := r.seats[seat]
		p.Ready = false
		p.PieceIDs = r.createArmyLocked(seat)
		r.randomizeSetupLocked(p)
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleRandomizeLocked(player *Player, refSeq int64) {
	if r.phase != PhaseSetup || player.Ready {
		r.sendErrorLocked(player.Token, "bad_phase", "当前不能随机布阵", refSeq)
		return
	}
	r.randomizeSetupLocked(player)
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handlePlaceLocked(player *Player, msg protocol.ClientMessage) {
	if r.phase != PhaseSetup || player.Ready {
		r.sendErrorLocked(player.Token, "bad_phase", "当前不能布阵", msg.Seq)
		return
	}
	piece := r.pieces[msg.PieceID]
	if piece.ID == 0 || piece.Owner != player.Seat {
		r.sendErrorLocked(player.Token, "bad_piece", "棋子不存在", msg.Seq)
		return
	}
	to := game.Pos{Row: msg.Row, Col: msg.Col}
	if err := game.ValidateSetupCellForMode(r.mode, player.Seat, piece.Rank, to); err != nil {
		r.sendErrorLocked(player.Token, "bad_setup", err.Error(), msg.Seq)
		return
	}
	from := r.placements[piece.ID]
	for _, id := range player.PieceIDs {
		if id != piece.ID && r.placements[id] == to {
			other := r.pieces[id]
			if err := game.ValidateSetupCellForMode(r.mode, player.Seat, other.Rank, from); err != nil {
				r.sendErrorLocked(player.Token, "bad_swap", "这两个棋子不能互换位置", msg.Seq)
				return
			}
			r.placements[id] = from
			r.placements[piece.ID] = to
			r.broadcastRoomAndViewsLocked()
			return
		}
	}
	r.placements[piece.ID] = to
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleSubmitLocked(player *Player, refSeq int64) {
	if r.phase != PhaseSetup {
		r.sendErrorLocked(player.Token, "bad_phase", "当前不能提交布阵", refSeq)
		return
	}
	pieces := make([]game.Piece, 0, len(player.PieceIDs))
	for _, id := range player.PieceIDs {
		pieces = append(pieces, r.pieces[id])
	}
	if err := game.ValidateSetupForMode(r.mode, player.Seat, pieces, r.placements); err != nil {
		r.sendErrorLocked(player.Token, "bad_setup", err.Error(), refSeq)
		return
	}
	player.Ready = true
	if r.allReadyLocked() {
		r.startPlayingLocked()
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleMoveLocked(player *Player, msg protocol.ClientMessage) {
	if r.phase != PhasePlaying || r.state == nil {
		r.sendErrorLocked(player.Token, "bad_phase", "当前不能走棋", msg.Seq)
		return
	}
	if r.request != nil {
		r.sendErrorLocked(player.Token, "request_pending", "有未处理的请求，请先回应", msg.Seq)
		return
	}
	if piece := r.state.Pieces[msg.PieceID]; piece.ID == 0 || piece.Owner != player.Seat {
		r.sendErrorLocked(player.Token, "bad_piece", "只能移动自己的棋子", msg.Seq)
		return
	}
	if msg.From == (game.Pos{}) {
		if pos, ok := r.state.Positions[msg.PieceID]; ok {
			msg.From = pos
		}
	}
	next, events, err := game.ApplyMove(r.state, game.Move{PieceID: msg.PieceID, From: msg.From, To: msg.To, Path: msg.Path})
	if err != nil {
		r.sendErrorLocked(player.Token, "illegal_move", err.Error(), msg.Seq)
		return
	}
	r.state = next
	if next.Phase == game.Ended {
		r.phase = PhaseEnded
		r.cancelTurnTimerLocked()
	} else {
		r.startTurnTimerLocked()
	}
	for _, event := range events {
		ev := event
		r.broadcastLocked(protocol.ServerMessage{Type: "event." + string(event.Type), Event: &ev})
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleSkipLocked(player *Player, refSeq int64) {
	if r.phase != PhasePlaying || r.state == nil {
		r.sendErrorLocked(player.Token, "bad_phase", "当前不能跳过", refSeq)
		return
	}
	if r.request != nil {
		r.sendErrorLocked(player.Token, "request_pending", "有未处理的请求，请先回应", refSeq)
		return
	}
	if r.state.Turn != player.Seat {
		r.sendErrorLocked(player.Token, "not_your_turn", "现在不是你的回合", refSeq)
		return
	}
	if r.state.Eliminated[player.Seat] {
		r.sendErrorLocked(player.Token, "eliminated", "你已被淘汰", refSeq)
		return
	}
	if r.skips[player.Seat] >= protocol.MaxSkipsPerPlayer {
		r.sendErrorLocked(player.Token, "skip_exhausted", "已达跳过上限", refSeq)
		return
	}
	r.skips[player.Seat]++
	if r.skips[player.Seat] >= protocol.MaxSkipsPerPlayer {
		r.broadcastNoticeLocked(player.Seat, "跳过次数用尽，判负")
		r.surrenderTeamLocked(player.Seat)
		return
	}
	r.advanceTurnAfterSkipLocked()
	r.broadcastNoticeLocked(player.Seat, "跳过回合")
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) advanceTurnAfterSkipLocked() {
	if r.state == nil {
		return
	}
	next := r.state.Clone()
	next.AdvanceTurn()
	r.state = next
	r.startTurnTimerLocked()
}

func (r *Room) onTurnTimeout(seat game.Seat, epoch int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if epoch != r.turnEpoch {
		return
	}
	r.onTurnTimeoutLocked(seat)
}

func (r *Room) onTurnTimeoutLocked(seat game.Seat) {
	if r.phase != PhasePlaying || r.state == nil || r.state.Turn != seat {
		return
	}
	if r.request != nil {
		return
	}
	if r.skips[seat] >= protocol.MaxSkipsPerPlayer {
		r.broadcastNoticeLocked(seat, "超时且跳过次数用尽，自动认输")
		r.surrenderTeamLocked(seat)
		return
	}
	r.skips[seat]++
	r.advanceTurnAfterSkipLocked()
	r.broadcastNoticeLocked(seat, "超时跳过回合")
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) startTurnTimerLocked() {
	r.cancelTurnTimerLocked()
	if r.phase != PhasePlaying || r.state == nil {
		return
	}
	if r.request != nil {
		return
	}
	if r.state.Eliminated[r.state.Turn] {
		return
	}
	limit := r.timeControl.MoveSeconds
	if limit <= 0 {
		limit = 15
	}
	duration := time.Duration(limit) * time.Second
	r.turnDeadline = time.Now().Add(duration)
	r.turnEpoch++
	epoch := r.turnEpoch
	seat := r.state.Turn
	r.turnTimer = time.AfterFunc(duration, func() {
		r.onTurnTimeout(seat, epoch)
	})
}

func (r *Room) cancelTurnTimerLocked() {
	if r.turnTimer != nil {
		r.turnTimer.Stop()
		r.turnTimer = nil
	}
	r.turnDeadline = time.Time{}
}

func (r *Room) eliminateSeatLocked(seat game.Seat) {
	if r.state == nil {
		return
	}
	r.state.Eliminate(seat)
	if winners := r.state.CheckWinner(); winners != nil {
		r.state.Phase = game.Ended
		r.state.Winner = winners
		r.phase = PhaseEnded
		r.cancelTurnTimerLocked()
		r.request = nil
		r.broadcastRoomAndViewsLocked()
		return
	}
	if r.state.Turn == seat {
		next := r.state.Clone()
		next.AdvanceTurn()
		r.state = next
	}
	r.startTurnTimerLocked()
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleRequestLocked(player *Player, kind string, refSeq int64) {
	if r.phase != PhasePlaying || r.state == nil {
		r.sendErrorLocked(player.Token, "bad_phase", "对局未在进行", refSeq)
		return
	}
	if kind != "tie" && kind != "surrender" {
		r.sendErrorLocked(player.Token, "bad_request", "未知的请求类型", refSeq)
		return
	}
	if r.request != nil {
		r.sendErrorLocked(player.Token, "request_pending", "已有进行中的请求", refSeq)
		return
	}
	if r.state.Eliminated[player.Seat] {
		r.sendErrorLocked(player.Token, "eliminated", "你已被淘汰，无法发起请求", refSeq)
		return
	}
	r.cancelTurnTimerLocked()
	if r.mode == game.ModeJunqi {
		if kind == "surrender" {
			r.broadcastNoticeLocked(player.Seat, "选择投降")
			r.surrenderTeamLocked(player.Seat)
			return
		}
		r.request = &protocol.PendingRequest{Kind: kind, From: player.Seat, Stage: "rivals"}
		r.broadcastNoticeLocked(player.Seat, "发起和棋请求，等待对方回应")
		r.broadcastRoomAndViewsLocked()
		return
	}

	partner := player.Seat.Partner()
	partnerAlive := !r.state.Eliminated[partner] && r.seats[partner] != nil
	if partnerAlive {
		r.request = &protocol.PendingRequest{Kind: kind, From: player.Seat, Stage: "teammate"}
		label := requestLabel(kind)
		r.broadcastNoticeLocked(player.Seat, "发起"+label+"请求，等待队友支持")
	} else {
		if kind == "surrender" {
			r.broadcastNoticeLocked(player.Seat, "选择投降")
			r.surrenderTeamLocked(player.Seat)
			return
		}
		r.request = &protocol.PendingRequest{Kind: kind, From: player.Seat, Stage: "rivals"}
		r.broadcastNoticeLocked(player.Seat, "发起和棋请求，等待对方回应")
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleRequestRespondLocked(player *Player, msg protocol.ClientMessage) {
	if r.request == nil {
		r.sendErrorLocked(player.Token, "no_request", "当前没有待回应的请求", msg.Seq)
		return
	}
	if msg.Kind != "" && msg.Kind != r.request.Kind {
		r.sendErrorLocked(player.Token, "request_mismatch", "请求已变化", msg.Seq)
		return
	}
	if player.Seat == r.request.From {
		r.sendErrorLocked(player.Token, "self_response", "不能回应自己发起的请求", msg.Seq)
		return
	}
	switch r.request.Stage {
	case "teammate":
		if r.mode == game.ModeJunqi {
			r.sendErrorLocked(player.Token, "bad_stage", "单挑模式没有队友回应阶段", msg.Seq)
			return
		}
		if player.Seat != r.request.From.Partner() {
			r.sendErrorLocked(player.Token, "not_teammate", "需要队友回应", msg.Seq)
			return
		}
		r.respondTeammateLocked(player, msg.Accept)
	case "rivals":
		if r.sameSideLocked(player.Seat, r.request.From) {
			r.sendErrorLocked(player.Token, "not_rival", "需要对方回应", msg.Seq)
			return
		}
		if r.state != nil && r.state.Eliminated[player.Seat] {
			r.sendErrorLocked(player.Token, "eliminated", "已被淘汰，无法回应", msg.Seq)
			return
		}
		r.respondRivalLocked(player, msg.Accept)
	default:
		r.sendErrorLocked(player.Token, "bad_stage", "请求阶段错误", msg.Seq)
	}
}

func (r *Room) respondTeammateLocked(player *Player, accept bool) {
	req := r.request
	label := requestLabel(req.Kind)
	if !accept {
		r.broadcastNoticeLocked(player.Seat, "拒绝了"+label+"请求")
		r.request = nil
		r.startTurnTimerLocked()
		r.broadcastRoomAndViewsLocked()
		return
	}
	r.broadcastNoticeLocked(player.Seat, "支持"+label+"请求")
	if req.Kind == "surrender" {
		r.surrenderTeamLocked(req.From)
		return
	}
	r.request = &protocol.PendingRequest{Kind: req.Kind, From: req.From, Stage: "rivals"}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) respondRivalLocked(player *Player, accept bool) {
	req := r.request
	label := requestLabel(req.Kind)
	if !accept {
		r.broadcastNoticeLocked(player.Seat, "拒绝了"+label+"请求")
		r.request = nil
		r.startTurnTimerLocked()
		r.broadcastRoomAndViewsLocked()
		return
	}
	for _, s := range req.Acks {
		if s == player.Seat {
			r.sendErrorLocked(player.Token, "already_acked", "你已回应", 0)
			return
		}
	}
	req.Acks = append(req.Acks, player.Seat)
	r.broadcastNoticeLocked(player.Seat, "同意"+label+"请求")
	if r.allRivalsAckedLocked(req) {
		if req.Kind == "tie" {
			r.declareDrawLocked()
			return
		}
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleRequestCancelLocked(player *Player, refSeq int64) {
	if r.request == nil {
		return
	}
	if r.request.From != player.Seat {
		r.sendErrorLocked(player.Token, "not_owner", "只能取消自己的请求", refSeq)
		return
	}
	r.broadcastNoticeLocked(player.Seat, "取消"+requestLabel(r.request.Kind)+"请求")
	r.request = nil
	r.startTurnTimerLocked()
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) allRivalsAckedLocked(req *protocol.PendingRequest) bool {
	if r.state == nil {
		return false
	}
	for _, seat := range game.ActiveSeats(r.mode) {
		if r.sameSideLocked(seat, req.From) {
			continue
		}
		if r.state.Eliminated[seat] {
			continue
		}
		if r.seats[seat] == nil {
			continue
		}
		acked := false
		for _, s := range req.Acks {
			if s == seat {
				acked = true
				break
			}
		}
		if !acked {
			return false
		}
	}
	return true
}

func (r *Room) surrenderTeamLocked(initiator game.Seat) {
	if r.state == nil {
		return
	}
	r.state.Eliminate(initiator)
	if r.mode != game.ModeJunqi {
		r.state.Eliminate(initiator.Partner())
	}
	winners := r.state.CheckWinner()
	if winners == nil {
		winners = []game.Seat{}
	}
	r.state.Phase = game.Ended
	r.state.Winner = winners
	r.phase = PhaseEnded
	r.request = nil
	r.cancelTurnTimerLocked()
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) sameSideLocked(a, b game.Seat) bool {
	if r.mode == game.ModeJunqi {
		return a == b
	}
	return a.SameTeam(b)
}

func (r *Room) declareDrawLocked() {
	if r.state == nil {
		return
	}
	r.state.Phase = game.Ended
	r.state.Winner = []game.Seat{}
	r.phase = PhaseEnded
	r.request = nil
	r.cancelTurnTimerLocked()
	r.broadcastRoomAndViewsLocked()
}

func requestLabel(kind string) string {
	if kind == "tie" {
		return "和棋"
	}
	if kind == "surrender" {
		return "投降"
	}
	return kind
}

func (r *Room) broadcastNoticeLocked(from game.Seat, text string) {
	chat := &protocol.ChatMessage{
		From:    from,
		Channel: ChannelAll,
		Text:    "[系统] " + text,
		TS:      time.Now().UnixMilli(),
	}
	if p := r.seats[from]; p != nil {
		chat.Name = p.Name
	}
	r.broadcastLocked(protocol.ServerMessage{Type: "chat.msg", Chat: chat})
}

func (r *Room) handleConcedeLocked(player *Player) {
	if r.phase != PhasePlaying || r.state == nil {
		return
	}
	r.state.Eliminate(player.Seat)
	if winners := r.state.CheckWinner(); winners != nil {
		r.state.Phase = game.Ended
		r.state.Winner = winners
		r.phase = PhaseEnded
		r.cancelTurnTimerLocked()
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleChatLocked(player *Player, msg protocol.ClientMessage) {
	channel := msg.Channel
	if channel == "" {
		channel = ChannelAll
	}
	if channel == ChannelTeam && !r.allowTeamChat {
		r.sendErrorLocked(player.Token, "team_chat_disabled", "队伍聊天已关闭", msg.Seq)
		return
	}
	if channel == ChannelTeam && r.mode == game.ModeJunqi {
		r.sendErrorLocked(player.Token, "team_chat_disabled", "单挑模式没有队伍聊天", msg.Seq)
		return
	}
	text := strings.TrimSpace(msg.Text)
	if len([]rune(text)) > 200 {
		r.sendErrorLocked(player.Token, "chat_too_long", "聊天最多 200 字", msg.Seq)
		return
	}
	if text == "" && msg.Emote == "" {
		return
	}
	if !r.allowChatLocked(player.Token) {
		r.sendErrorLocked(player.Token, "rate_limited", "发言太快了", msg.Seq)
		return
	}
	chat := &protocol.ChatMessage{
		From:    player.Seat,
		Name:    player.Name,
		Channel: channel,
		Text:    text,
		Emote:   msg.Emote,
		TS:      time.Now().UnixMilli(),
	}
	if channel == ChannelTeam {
		r.sendSeatLocked(player.Seat, protocol.ServerMessage{Type: "chat.msg", Chat: chat})
		r.sendSeatLocked(player.Seat.Partner(), protocol.ServerMessage{Type: "chat.msg", Chat: chat})
		return
	}
	r.broadcastLocked(protocol.ServerMessage{Type: "chat.msg", Chat: chat})
}

func (r *Room) createArmyLocked(seat game.Seat) []game.PieceID {
	var ids []game.PieceID
	for rank, count := range game.StandardArmy {
		for i := 0; i < count; i++ {
			id := r.nextPieceID
			r.nextPieceID++
			r.pieces[id] = game.Piece{ID: id, Owner: seat, Rank: rank, Alive: true}
			ids = append(ids, id)
		}
	}
	return ids
}

func (r *Room) randomizeSetupLocked(player *Player) {
	for _, id := range player.PieceIDs {
		delete(r.placements, id)
	}

	ids := orderedSetupIDs(player.PieceIDs, r.pieces)
	used := map[game.Pos]bool{}
	for _, id := range ids {
		piece := r.pieces[id]
		slots := setupSlots(r.mode, player.Seat, piece.Rank, used)
		if len(slots) == 0 {
			r.placements = fallbackSetup(r.mode, player.Seat, player.PieceIDs, r.pieces, r.placements)
			return
		}
		pos := slots[mrand.Intn(len(slots))]
		used[pos] = true
		r.placements[id] = pos
	}
}

func orderedSetupIDs(ids []game.PieceID, pieces map[game.PieceID]game.Piece) []game.PieceID {
	out := append([]game.PieceID(nil), ids...)
	mrand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	priority := func(rank game.Rank) int {
		switch rank {
		case game.Flag:
			return 0
		case game.Mine:
			return 1
		case game.Bomb:
			return 2
		default:
			return 3
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return priority(pieces[out[i]].Rank) < priority(pieces[out[j]].Rank)
	})
	return out
}

func fallbackSetup(mode game.GameMode, seat game.Seat, ids []game.PieceID, pieces map[game.PieceID]game.Piece, placements map[game.PieceID]game.Pos) map[game.PieceID]game.Pos {
	for _, id := range ids {
		delete(placements, id)
	}
	used := map[game.Pos]bool{}
	for _, id := range orderedSetupIDs(ids, pieces) {
		piece := pieces[id]
		slots := setupSlots(mode, seat, piece.Rank, used)
		if len(slots) == 0 {
			continue
		}
		pos := slots[0]
		used[pos] = true
		placements[id] = pos
	}
	return placements
}

func setupSlots(mode game.GameMode, seat game.Seat, rank game.Rank, used map[game.Pos]bool) []game.Pos {
	var slots []game.Pos
	for row := 0; row < game.BoardSize; row++ {
		for col := 0; col < game.BoardSize; col++ {
			pos := game.Pos{Row: row, Col: col}
			if used[pos] {
				continue
			}
			if err := game.ValidateSetupCellForMode(mode, seat, rank, pos); err == nil {
				slots = append(slots, pos)
			}
		}
	}
	return slots
}

func (r *Room) startPlayingLocked() {
	pieces := make([]game.Piece, 0, len(r.pieces))
	for _, piece := range r.pieces {
		pieces = append(pieces, piece)
	}
	state, err := game.NewStateForMode(r.mode, pieces, r.placements, game.North)
	if err != nil {
		r.broadcastLocked(protocol.ServerMessage{Type: "error", Error: &protocol.ErrorMessage{Code: "start_failed", Message: err.Error()}})
		return
	}
	r.state = state
	r.phase = PhasePlaying
	r.skips = map[game.Seat]int{}
	r.request = nil
	r.startTurnTimerLocked()
}

func (r *Room) allReadyLocked() bool {
	for _, seat := range game.ActiveSeats(r.mode) {
		p := r.seats[seat]
		if p == nil || !p.Ready {
			return false
		}
	}
	return true
}

func (r *Room) stateWinnerLocked() []game.Seat {
	if r.mode == game.ModeJunqi {
		if r.state.Eliminated[game.North] {
			return []game.Seat{game.South}
		}
		if r.state.Eliminated[game.South] {
			return []game.Seat{game.North}
		}
		return nil
	}
	nsLost := r.state.Eliminated[game.North] && r.state.Eliminated[game.South]
	ewLost := r.state.Eliminated[game.East] && r.state.Eliminated[game.West]
	if nsLost {
		return []game.Seat{game.East, game.West}
	}
	if ewLost {
		return []game.Seat{game.North, game.South}
	}
	return nil
}

func (r *Room) viewForLocked(seat game.Seat) *game.ClientView {
	if r.phase == PhaseLobby {
		return nil
	}
	if r.state != nil {
		view := r.state.ViewFor(seat)
		return &view
	}
	pieces := make([]game.Piece, 0, len(r.pieces))
	for _, piece := range r.pieces {
		pieces = append(pieces, piece)
	}
	state, err := game.NewStateForMode(r.mode, pieces, r.placements, game.North)
	if err != nil {
		return nil
	}
	state.Phase = game.Setup
	view := state.ViewFor(seat)
	return &view
}

func (r *Room) snapshotLocked(viewer game.Seat) *protocol.RoomSnapshot {
	snap := &protocol.RoomSnapshot{
		Code:          r.Code,
		Mode:          r.mode,
		Phase:         r.phase,
		HostSeat:      r.host,
		AllowTeamChat: r.allowTeamChat,
		TimeControl:   r.timeControl,
		Turn:          game.North,
		View:          r.viewForLocked(viewer),
		MaxSkips:      protocol.MaxSkipsPerPlayer,
		MoveLimitSec:  r.timeControl.MoveSeconds,
	}
	if len(r.skips) > 0 {
		skipsCopy := make(map[game.Seat]int, len(r.skips))
		for k, v := range r.skips {
			skipsCopy[k] = v
		}
		snap.Skips = skipsCopy
	}
	if r.phase == PhasePlaying && !r.turnDeadline.IsZero() {
		snap.MoveDeadlineMs = r.turnDeadline.UnixMilli()
	}
	if r.request != nil {
		req := *r.request
		req.Acks = append([]game.Seat(nil), r.request.Acks...)
		snap.Request = &req
	}
	if r.state != nil {
		snap.Turn = r.state.Turn
		snap.Winner = append([]game.Seat(nil), r.state.Winner...)
		if len(r.state.Eliminated) > 0 {
			elim := map[game.Seat]bool{}
			for s, v := range r.state.Eliminated {
				if v {
					elim[s] = true
				}
			}
			if len(elim) > 0 {
				snap.Eliminated = elim
			}
		}
	}
	for _, seat := range game.ActiveSeats(r.mode) {
		info := protocol.SeatInfo{Seat: seat}
		if p := r.seats[seat]; p != nil {
			info.Name = p.Name
			info.Connected = p.Connected
			info.Host = p.Host
			info.Ready = p.Ready
			if p.Ready {
				snap.Ready = append(snap.Ready, seat)
			}
		}
		snap.Seats = append(snap.Seats, info)
	}
	return snap
}

func (r *Room) broadcastRoomAndViewsLocked() {
	r.broadcastRoomLocked()
	for _, player := range r.players {
		r.sendLocked(player.Token, protocol.ServerMessage{Type: "view", View: r.viewForLocked(player.Seat)})
	}
}

func (r *Room) broadcastRoomLocked() {
	for _, player := range r.players {
		r.sendLocked(player.Token, protocol.ServerMessage{Type: "room.state", Room: r.snapshotLocked(player.Seat)})
	}
}

func (r *Room) broadcastLocked(msg protocol.ServerMessage) {
	for token := range r.connections {
		r.sendLocked(token, msg)
	}
}

func (r *Room) sendSeatLocked(seat game.Seat, msg protocol.ServerMessage) {
	if p := r.seats[seat]; p != nil {
		r.sendLocked(p.Token, msg)
	}
}

func (r *Room) sendErrorLocked(token, code, message string, refSeq int64) {
	r.sendLocked(token, protocol.ServerMessage{
		Type:  "error",
		Error: &protocol.ErrorMessage{Code: code, Message: message, RefSeq: refSeq},
	})
}

func (r *Room) sendLocked(token string, msg protocol.ServerMessage) {
	ch := r.connections[token]
	if ch == nil {
		return
	}
	r.serverSeq++
	msg.Seq = r.serverSeq
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case ch <- data:
	default:
	}
}

func (r *Room) allowChatLocked(token string) bool {
	now := time.Now()
	window := r.chatWindow[token]
	kept := window[:0]
	for _, ts := range window {
		if now.Sub(ts) <= 10*time.Second {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= 5 {
		r.chatWindow[token] = kept
		return false
	}
	r.chatWindow[token] = append(kept, now)
	return true
}

func firstOpenSeat(mode game.GameMode, seats map[game.Seat]*Player) game.Seat {
	for _, seat := range game.ActiveSeats(mode) {
		if seats[seat] == nil {
			return seat
		}
	}
	return game.North
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "玩家"
	}
	runes := []rune(name)
	if len(runes) > 16 {
		name = string(runes[:16])
	}
	return name
}

func newToken() string {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw[:])
}

func NewCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := 0; i < 6; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "ROOM01"
		}
		b.WriteByte(alphabet[n.Int64()])
	}
	return b.String()
}
