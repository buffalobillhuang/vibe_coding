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
}

func New(code string) *Room {
	return &Room{
		Code:          code,
		phase:         PhaseLobby,
		allowTeamChat: true,
		timeControl:   protocol.DefaultTimeControl(),
		players:       map[string]*Player{},
		seats:         map[game.Seat]*Player{},
		connections:   map[string]chan []byte{},
		pieces:        map[game.PieceID]game.Piece{},
		placements:    map[game.PieceID]game.Pos{},
		nextPieceID:   1,
		chatWindow:    map[string][]time.Time{},
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
	if len(r.seats) >= 4 {
		return nil, ErrRoomFull
	}

	seat := firstOpenSeat(r.seats)
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

func (r *Room) ConfigureInitial(tc *protocol.TimeControl, allowTeamChat *bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.phase != PhaseLobby {
		return
	}
	if tc != nil {
		r.timeControl = *tc
	}
	if allowTeamChat != nil {
		r.allowTeamChat = *allowTeamChat
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
	if msg.TimeControl != nil {
		r.timeControl = *msg.TimeControl
	}
	if msg.AllowTeamChat != nil {
		r.allowTeamChat = *msg.AllowTeamChat
	}
	r.broadcastRoomLocked()
}

func (r *Room) handleStartLocked(player *Player, refSeq int64) {
	if !player.Host {
		r.sendErrorLocked(player.Token, "forbidden", "只有房主可以开始", refSeq)
		return
	}
	if len(r.seats) != 4 {
		r.sendErrorLocked(player.Token, "not_ready", "需要四名玩家", refSeq)
		return
	}
	if r.phase != PhaseLobby {
		return
	}

	r.phase = PhaseSetup
	r.pieces = map[game.PieceID]game.Piece{}
	r.placements = map[game.PieceID]game.Pos{}
	for _, seat := range game.Seats {
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
	if err := game.ValidateSetupCell(player.Seat, piece.Rank, to); err != nil {
		r.sendErrorLocked(player.Token, "bad_setup", err.Error(), msg.Seq)
		return
	}
	from := r.placements[piece.ID]
	for _, id := range player.PieceIDs {
		if id != piece.ID && r.placements[id] == to {
			other := r.pieces[id]
			if err := game.ValidateSetupCell(player.Seat, other.Rank, from); err != nil {
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
	if err := game.ValidateSetup(player.Seat, pieces, r.placements); err != nil {
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
	}
	for _, event := range events {
		ev := event
		r.broadcastLocked(protocol.ServerMessage{Type: "event." + string(event.Type), Event: &ev})
	}
	r.broadcastRoomAndViewsLocked()
}

func (r *Room) handleConcedeLocked(player *Player) {
	if r.phase != PhasePlaying || r.state == nil {
		return
	}
	r.state.Eliminated[player.Seat] = true
	if winners := r.state.Winner; len(winners) == 0 {
		if calculated := r.state.Clone().Winner; len(calculated) > 0 {
			r.state.Winner = calculated
		}
	}
	if winners := r.stateWinnerLocked(); winners != nil {
		r.state.Phase = game.Ended
		r.state.Winner = winners
		r.phase = PhaseEnded
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
		slots := setupSlots(player.Seat, piece.Rank, used)
		if len(slots) == 0 {
			r.placements = fallbackSetup(player.Seat, player.PieceIDs, r.pieces, r.placements)
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

func fallbackSetup(seat game.Seat, ids []game.PieceID, pieces map[game.PieceID]game.Piece, placements map[game.PieceID]game.Pos) map[game.PieceID]game.Pos {
	for _, id := range ids {
		delete(placements, id)
	}
	used := map[game.Pos]bool{}
	for _, id := range orderedSetupIDs(ids, pieces) {
		piece := pieces[id]
		slots := setupSlots(seat, piece.Rank, used)
		if len(slots) == 0 {
			continue
		}
		pos := slots[0]
		used[pos] = true
		placements[id] = pos
	}
	return placements
}

func setupSlots(seat game.Seat, rank game.Rank, used map[game.Pos]bool) []game.Pos {
	var slots []game.Pos
	for row := 0; row < game.BoardSize; row++ {
		for col := 0; col < game.BoardSize; col++ {
			pos := game.Pos{Row: row, Col: col}
			if used[pos] {
				continue
			}
			if err := game.ValidateSetupCell(seat, rank, pos); err == nil {
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
	state, err := game.NewState(pieces, r.placements, game.North)
	if err != nil {
		r.broadcastLocked(protocol.ServerMessage{Type: "error", Error: &protocol.ErrorMessage{Code: "start_failed", Message: err.Error()}})
		return
	}
	r.state = state
	r.phase = PhasePlaying
}

func (r *Room) allReadyLocked() bool {
	for _, seat := range game.Seats {
		p := r.seats[seat]
		if p == nil || !p.Ready {
			return false
		}
	}
	return true
}

func (r *Room) stateWinnerLocked() []game.Seat {
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
	state, err := game.NewState(pieces, r.placements, game.North)
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
		Phase:         r.phase,
		HostSeat:      r.host,
		AllowTeamChat: r.allowTeamChat,
		TimeControl:   r.timeControl,
		Turn:          game.North,
		View:          r.viewForLocked(viewer),
	}
	if r.state != nil {
		snap.Turn = r.state.Turn
		snap.Winner = append([]game.Seat(nil), r.state.Winner...)
	}
	for _, seat := range game.Seats {
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

func firstOpenSeat(seats map[game.Seat]*Player) game.Seat {
	for _, seat := range game.Seats {
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
