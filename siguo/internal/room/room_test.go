package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"siguo/internal/game"
	"siguo/internal/protocol"
)

func TestSetupPlaceSwapsOwnPieces(t *testing.T) {
	r := New("ABC123")
	p, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	p.PieceIDs = []game.PieceID{1, 2}
	r.phase = PhaseSetup
	r.pieces[1] = game.Piece{ID: 1, Owner: p.Seat, Rank: game.Engineer, Alive: true}
	r.pieces[2] = game.Piece{ID: 2, Owner: p.Seat, Rank: game.Commander, Alive: true}
	r.placements[1] = game.Pos{Row: 1, Col: 8}
	r.placements[2] = game.Pos{Row: 1, Col: 9}

	sendTestMessage(t, r, p.Token, protocol.ClientMessage{
		Type:    "setup.place",
		Seq:     1,
		PieceID: 1,
		Row:     1,
		Col:     9,
	})

	if got := r.placements[1]; got != (game.Pos{Row: 1, Col: 9}) {
		t.Fatalf("piece 1 position = %v, want {1 9}", got)
	}
	if got := r.placements[2]; got != (game.Pos{Row: 1, Col: 8}) {
		t.Fatalf("piece 2 position = %v, want {1 8}", got)
	}
}

func TestActiveReportsSetupAndPlayingOnly(t *testing.T) {
	r := New("ABC123")
	if r.Active() {
		t.Fatal("new lobby room should not be active")
	}
	r.phase = PhaseSetup
	if !r.Active() {
		t.Fatal("setup room should be active")
	}
	r.phase = PhasePlaying
	if !r.Active() {
		t.Fatal("playing room should be active")
	}
	r.phase = PhaseEnded
	if r.Active() {
		t.Fatal("ended room should not be active")
	}
}

func TestConnectViewerCapsAtMaxViewers(t *testing.T) {
	r := newPlayingRoom(t)
	var viewerIDs []string
	for i := 0; i < MaxViewers; i++ {
		_, id, err := r.ConnectViewer(fmt.Sprintf("viewer%d", i+1))
		if err != nil {
			t.Fatalf("ConnectViewer(%d) error = %v", i, err)
		}
		viewerIDs = append(viewerIDs, id)
	}
	if _, _, err := r.ConnectViewer("overflow-viewer"); !errors.Is(err, ErrViewersFull) {
		t.Fatalf("ConnectViewer() error = %v, want ErrViewersFull", err)
	}
	r.DisconnectViewer(viewerIDs[0])
	if _, _, err := r.ConnectViewer("replacement-viewer"); err != nil {
		t.Fatalf("ConnectViewer() after disconnect error = %v", err)
	}
}

func TestConnectViewerRejectsPlayerName(t *testing.T) {
	r := newPlayingRoom(t)
	if _, _, err := r.ConnectViewer(" north "); !errors.Is(err, ErrViewerNameConflict) {
		t.Fatalf("ConnectViewer(player name) error = %v, want ErrViewerNameConflict", err)
	}
	if _, _, err := r.ConnectViewer("   "); !errors.Is(err, ErrViewerNameRequired) {
		t.Fatalf("ConnectViewer(blank) error = %v, want ErrViewerNameRequired", err)
	}
	if _, _, err := r.ConnectViewer("observer"); err != nil {
		t.Fatalf("ConnectViewer(observer) error = %v", err)
	}
}

func TestConnectViewerAllowsLobbyWithTwoPlayers(t *testing.T) {
	r := New("ABC123")
	if _, err := r.Join("north", ""); err != nil {
		t.Fatalf("Join(north) error = %v", err)
	}
	if _, err := r.Join("south", ""); err != nil {
		t.Fatalf("Join(south) error = %v", err)
	}
	out, id, err := r.ConnectViewer("observer")
	if err != nil {
		t.Fatalf("ConnectViewer(observer) error = %v", err)
	}
	if out == nil || id == "" {
		t.Fatalf("ConnectViewer(observer) = %v, %q, want channel and id", out, id)
	}
}

func TestViewerCanSendPublicChat(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	northOut, _, err := r.Connect(north.Token)
	if err != nil {
		t.Fatalf("Connect(north) error = %v", err)
	}
	drainMessages(northOut)

	viewerOut, viewerID, err := r.ConnectViewer("observer")
	if err != nil {
		t.Fatalf("ConnectViewer(observer) error = %v", err)
	}
	drainMessages(viewerOut)
	drainMessages(northOut)

	sendViewerTestMessage(t, r, viewerID, protocol.ClientMessage{Type: "chat.send", Seq: 1, Channel: ChannelAll, Text: "hello board"})

	msg := nextServerMessage(t, northOut)
	if msg.Type != "chat.msg" || msg.Chat == nil {
		t.Fatalf("message = %+v, want chat.msg", msg)
	}
	if !msg.Chat.Viewer || msg.Chat.Name != "observer" || msg.Chat.Text != "hello board" {
		t.Fatalf("viewer chat = %+v, want public viewer chat", msg.Chat)
	}
	viewerEcho := nextServerMessage(t, viewerOut)
	if viewerEcho.Type != "chat.msg" || viewerEcho.Chat == nil || viewerEcho.Chat.Text != "hello board" {
		t.Fatalf("viewer echo = %+v, want echoed public chat", viewerEcho)
	}

	sendViewerTestMessage(t, r, viewerID, protocol.ClientMessage{Type: "chat.send", Seq: 2, Channel: ChannelTeam, Text: "nope"})
	msg = nextServerMessage(t, viewerOut)
	if msg.Type != "error" || msg.Error == nil || msg.Error.Code != "viewer_all_chat_only" {
		t.Fatalf("viewer team chat response = %+v, want viewer_all_chat_only error", msg)
	}
}

func TestBroadcastViewerNoticeUsesChatStream(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	northOut, _, err := r.Connect(north.Token)
	if err != nil {
		t.Fatalf("Connect(north) error = %v", err)
	}
	drainMessages(northOut)

	_, viewerID, err := r.ConnectViewer("observer")
	if err != nil {
		t.Fatalf("ConnectViewer(observer) error = %v", err)
	}
	drainMessages(northOut)

	r.BroadcastViewerNotice(r.ViewerName(viewerID), "加入观战")
	joined := nextServerMessage(t, northOut)
	if joined.Type != "chat.msg" || joined.Chat == nil || !strings.Contains(joined.Chat.Text, "observer加入观战") {
		t.Fatalf("joined notice = %+v, want observer join chat notice", joined)
	}

	r.BroadcastViewerNotice(r.ViewerName(viewerID), "离开观战")
	left := nextServerMessage(t, northOut)
	if left.Type != "chat.msg" || left.Chat == nil || !strings.Contains(left.Chat.Text, "observer离开观战") {
		t.Fatalf("left notice = %+v, want observer leave chat notice", left)
	}
}

func TestStartReportsFriendlyMessageWhenActiveRoomsAreFull(t *testing.T) {
	r := New("ABC123")
	r.SetActiveHooks(func() bool { return false }, nil)
	host, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(host) error = %v", err)
	}
	for _, name := range []string{"east", "south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	out, _, err := r.Connect(host.Token)
	if err != nil {
		t.Fatalf("Connect(host) error = %v", err)
	}
	drainMessages(out)

	sendTestMessage(t, r, host.Token, protocol.ClientMessage{Type: "room.start", Seq: 1})

	select {
	case data := <-out:
		var msg protocol.ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("Unmarshal(server message) error = %v", err)
		}
		if msg.Type != "error" || msg.Error == nil || msg.Error.Code != "active_rooms_full" {
			t.Fatalf("message = %+v, want active_rooms_full error", msg)
		}
		if !strings.Contains(msg.Error.Message, "所有客房已满") {
			t.Fatalf("error message = %q, missing friendly full-room message", msg.Error.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for full-room error")
	}
	if r.phase != PhaseLobby {
		t.Fatalf("phase = %s, want lobby after failed activation", r.phase)
	}
}

func TestJoinAssignsSystemNamesWhenBlank(t *testing.T) {
	r := New("ABC123")
	first, err := r.Join("", "")
	if err != nil {
		t.Fatalf("Join(blank first) error = %v", err)
	}
	second, err := r.Join("   ", "")
	if err != nil {
		t.Fatalf("Join(blank second) error = %v", err)
	}
	if first.Name != "玩家1" || second.Name != "玩家2" {
		t.Fatalf("names = %q/%q, want 玩家1/玩家2", first.Name, second.Name)
	}
	if _, err := r.Join("", first.Token); err != nil {
		t.Fatalf("Reconnect(blank) error = %v", err)
	}
	if first.Name != "玩家1" {
		t.Fatalf("blank reconnect renamed player to %q", first.Name)
	}
}

func TestReconnectDoesNotLetStaleDisconnectKillNewSocket(t *testing.T) {
	r := New("ABC123")
	player, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}

	first, _, err := r.Connect(player.Token)
	if err != nil {
		t.Fatalf("Connect(first) error = %v", err)
	}
	drainMessages(first)

	second, _, err := r.Connect(player.Token)
	if err != nil {
		t.Fatalf("Connect(second) error = %v", err)
	}
	drainMessages(second)

	select {
	case _, ok := <-first:
		if ok {
			t.Fatal("old connection should be closed when a new connection replaces it")
		}
	default:
		t.Fatal("old connection should already be closed")
	}

	r.Disconnect(player.Token, first)
	if !player.Connected {
		t.Fatal("stale disconnect should not mark the player offline")
	}

	r.mu.Lock()
	r.sendLocked(player.Token, protocol.ServerMessage{Type: "error", Error: &protocol.ErrorMessage{Code: "test", Message: "still-connected"}})
	r.mu.Unlock()

	select {
	case <-second:
	case <-time.After(time.Second):
		t.Fatal("replacement connection did not receive server messages after stale disconnect")
	}
	drainMessages(second)

	r.Disconnect(player.Token, second)
	if player.Connected {
		t.Fatal("active disconnect should mark the player offline")
	}
	select {
	case _, ok := <-second:
		if ok {
			t.Fatal("replacement connection should be closed on active disconnect")
		}
	default:
		t.Fatal("replacement connection should already be closed after active disconnect")
	}
}

func TestFullSiguoLobbyPlayersCanSwapSeats(t *testing.T) {
	r := New("ABC123")
	north, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(north) error = %v", err)
	}
	east, err := r.Join("east", "")
	if err != nil {
		t.Fatalf("Join(east) error = %v", err)
	}
	for _, name := range []string{"south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}

	target := game.East
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "seat.swap", Seq: 1, Seat: &target})

	if north.Seat != game.East || east.Seat != game.North {
		t.Fatalf("seats after swap = north:%s east:%s, want East/North", north.Seat, east.Seat)
	}
	if r.seats[game.East] != north || r.seats[game.North] != east {
		t.Fatalf("seat map not swapped")
	}
	if r.host != game.East {
		t.Fatalf("host seat = %s, want East after host swaps", r.host)
	}
	snap, err := r.SnapshotFor(north.Token)
	if err != nil {
		t.Fatalf("SnapshotFor(north) error = %v", err)
	}
	if snap.SelfSeat != game.East {
		t.Fatalf("snapshot selfSeat = %s, want East", snap.SelfSeat)
	}
}

func TestRoomMoveCanAttackEnemyPiece(t *testing.T) {
	r := New("ABC123")
	p, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	r.phase = PhasePlaying
	state, err := game.NewState(
		[]game.Piece{
			{ID: 1, Owner: game.North, Rank: game.Commander, Alive: true},
			{ID: 2, Owner: game.East, Rank: game.Engineer, Alive: true},
		},
		map[game.PieceID]game.Pos{
			1: {Row: 1, Col: 8},
			2: {Row: 2, Col: 8},
		},
		game.North,
	)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	r.state = state

	sendTestMessage(t, r, p.Token, protocol.ClientMessage{
		Type:    "move",
		Seq:     1,
		PieceID: 1,
		From:    game.Pos{Row: 1, Col: 8},
		To:      game.Pos{Row: 2, Col: 8},
	})

	if _, ok := r.state.Positions[2]; ok {
		t.Fatal("defender still has a board position after attack")
	}
	if got := r.state.Positions[1]; got != (game.Pos{Row: 2, Col: 8}) {
		t.Fatalf("attacker position = %v, want {2 8}", got)
	}
}

func TestStartCreatesVisibleSetupBoardForAllSeats(t *testing.T) {
	for i := 0; i < 25; i++ {
		r := New("ABC123")
		host, err := r.Join("north", "")
		if err != nil {
			t.Fatalf("Join(host) error = %v", err)
		}
		for _, name := range []string{"east", "south", "west"} {
			if _, err := r.Join(name, ""); err != nil {
				t.Fatalf("Join(%s) error = %v", name, err)
			}
		}

		sendTestMessage(t, r, host.Token, protocol.ClientMessage{Type: "room.start", Seq: 1})

		if r.phase != PhaseSetup {
			t.Fatalf("phase = %s, want setup", r.phase)
		}
		for _, seat := range game.Seats {
			player := r.seats[seat]
			pieces := make([]game.Piece, 0, len(player.PieceIDs))
			for _, id := range player.PieceIDs {
				pieces = append(pieces, r.pieces[id])
			}
			if err := game.ValidateSetup(seat, pieces, r.placements); err != nil {
				t.Fatalf("seat %s setup invalid: %v", seat, err)
			}
			view := r.viewForLocked(seat)
			if view == nil || len(view.Cells) == 0 {
				t.Fatalf("seat %s view is empty", seat)
			}
		}
		r.mu.Lock()
		r.cancelSetupTimerLocked()
		r.mu.Unlock()
	}
}

func TestSetupTimeoutAutoSubmitsAndStartsPlaying(t *testing.T) {
	r := New("ABC123")
	r.ConfigureInitial(game.ModeSiguo, &protocol.TimeControl{SetupSeconds: 3600, MoveSeconds: 15}, nil)
	host, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(host) error = %v", err)
	}
	for _, name := range []string{"east", "south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	sendTestMessage(t, r, host.Token, protocol.ClientMessage{Type: "room.start", Seq: 1})
	if r.phase != PhaseSetup {
		t.Fatalf("phase = %s, want setup", r.phase)
	}
	if r.setupTimer == nil || r.setupDeadline.IsZero() {
		t.Fatal("setup timer was not started")
	}
	epoch := r.setupEpoch
	r.onSetupTimeout(epoch)
	if r.phase != PhasePlaying {
		t.Fatalf("phase = %s, want playing after setup timeout", r.phase)
	}
	if r.state == nil {
		t.Fatal("state is nil after setup timeout")
	}
	if r.setupTimer != nil || !r.setupDeadline.IsZero() {
		t.Fatal("setup timer was not cleared after auto-start")
	}
	for _, seat := range game.ActiveSeats(r.mode) {
		if !r.seats[seat].Ready {
			t.Fatalf("seat %s was not auto-submitted", seat)
		}
	}
	r.mu.Lock()
	r.cancelTurnTimerLocked()
	r.mu.Unlock()
}

func TestAllSetupSubmissionsCancelSetupTimer(t *testing.T) {
	r := New("ABC123")
	r.ConfigureInitial(game.ModeSiguo, &protocol.TimeControl{SetupSeconds: 3600, MoveSeconds: 15}, nil)
	host, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(host) error = %v", err)
	}
	for _, name := range []string{"east", "south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	sendTestMessage(t, r, host.Token, protocol.ClientMessage{Type: "room.start", Seq: 1})
	if r.setupTimer == nil {
		t.Fatal("setup timer was not started")
	}
	seq := int64(2)
	for _, player := range r.seats {
		sendTestMessage(t, r, player.Token, protocol.ClientMessage{Type: "setup.submit", Seq: seq})
		seq++
	}
	if r.phase != PhasePlaying {
		t.Fatalf("phase = %s, want playing after all submissions", r.phase)
	}
	if r.setupTimer != nil || !r.setupDeadline.IsZero() {
		t.Fatal("setup timer was not cleared after all submissions")
	}
	r.mu.Lock()
	r.cancelTurnTimerLocked()
	r.mu.Unlock()
}

func TestSkipAdvancesTurnAndCounts(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	if r.state.Turn != game.North {
		t.Fatalf("turn = %v, want North", r.state.Turn)
	}
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "move.skip", Seq: 1})
	if r.state.Turn == game.North {
		t.Fatalf("turn did not advance after skip")
	}
	if r.skips[game.North] != 1 {
		t.Fatalf("north skips = %d, want 1", r.skips[game.North])
	}
}

func TestSkipBeyondLimitRejected(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	r.skips[game.North] = protocol.MaxSkipsPerPlayer
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "move.skip", Seq: 1})
	if r.skips[game.North] != protocol.MaxSkipsPerPlayer {
		t.Fatalf("skip count changed past limit")
	}
}

func TestFifthManualSkipConcedesSide(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	r.skips[game.North] = protocol.MaxSkipsPerPlayer - 1
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "move.skip", Seq: 1})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended after fifth skip", r.phase)
	}
	if len(r.state.Winner) != 2 || !(r.state.Winner[0].SameTeam(game.East) && r.state.Winner[1].SameTeam(game.East)) {
		t.Fatalf("winner = %v, want EW team", r.state.Winner)
	}
}

func TestTieRequestRequiresTeammateThenRivals(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	south := r.seats[game.South]
	east := r.seats[game.East]
	west := r.seats[game.West]

	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "request.tie", Seq: 1})
	if r.request == nil || r.request.Stage != "teammate" || r.request.Kind != "tie" {
		t.Fatalf("expected teammate-stage tie request, got %+v", r.request)
	}
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: true})
	if r.request == nil || r.request.Stage != "rivals" {
		t.Fatalf("expected rivals-stage tie after partner support, got %+v", r.request)
	}
	sendTestMessage(t, r, east.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: true})
	if r.request == nil {
		t.Fatalf("request resolved before all rivals acked")
	}
	sendTestMessage(t, r, west.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: true})
	if r.phase != PhaseEnded || r.state.Phase != game.Ended {
		t.Fatalf("phase = %s, want ended after tie", r.phase)
	}
	if len(r.state.Winner) != 0 {
		t.Fatalf("winner = %v, want empty for draw", r.state.Winner)
	}
}

func TestSurrenderRequestEndsAfterTeammateSupport(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	south := r.seats[game.South]

	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "request.surrender", Seq: 1})
	if r.request == nil || r.request.Kind != "surrender" {
		t.Fatalf("expected surrender request pending")
	}
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: true})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended after surrender", r.phase)
	}
	if !r.state.Eliminated[game.North] || !r.state.Eliminated[game.South] {
		t.Fatalf("expected NS team eliminated after surrender")
	}
	if len(r.state.Winner) != 2 || !(r.state.Winner[0].SameTeam(game.East) && r.state.Winner[1].SameTeam(game.East)) {
		t.Fatalf("winner = %v, want EW team", r.state.Winner)
	}
}

func TestTeamWinsWhenBothRivalsEliminated(t *testing.T) {
	r := newPlayingRoom(t)
	r.mu.Lock()
	r.eliminateSeatLocked(game.East)
	if r.phase == PhaseEnded {
		r.mu.Unlock()
		t.Fatalf("game should still be playing after only one rival eliminated")
	}
	r.eliminateSeatLocked(game.West)
	r.mu.Unlock()
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended after both rivals eliminated", r.phase)
	}
	if len(r.state.Winner) != 2 || r.state.Winner[0] != game.North || r.state.Winner[1] != game.South {
		t.Fatalf("winners = %v, want NS team", r.state.Winner)
	}
}

func TestActiveSummaryIncludesPopulatedLobbyForWatching(t *testing.T) {
	r := New("ABC123")
	if _, ok := r.ActiveSummary(); ok {
		t.Fatal("empty lobby should not appear in watch-room summary")
	}
	if _, err := r.Join("north", ""); err != nil {
		t.Fatalf("Join(north) error = %v", err)
	}
	if _, ok := r.ActiveSummary(); ok {
		t.Fatal("single-seat lobby should not appear in watch-room summary")
	}
	if _, err := r.Join("east", ""); err != nil {
		t.Fatalf("Join(east) error = %v", err)
	}
	info, ok := r.ActiveSummary()
	if !ok {
		t.Fatal("populated lobby should appear in watch-room summary")
	}
	if info.Phase != PhaseLobby {
		t.Fatalf("summary phase = %s, want lobby", info.Phase)
	}
}

func TestJunqiRoomStartsWithTwoPlayers(t *testing.T) {
	r := New("DUEL01")
	r.ConfigureInitial(game.ModeJunqi, nil, nil)
	north, err := r.Join("north", "")
	if err != nil {
		t.Fatalf("Join(north) error = %v", err)
	}
	south, err := r.Join("south", "")
	if err != nil {
		t.Fatalf("Join(south) error = %v", err)
	}
	if north.Seat != game.North || south.Seat != game.South {
		t.Fatalf("seats = %s/%s, want North/South", north.Seat, south.Seat)
	}
	if _, err := r.Join("extra", ""); err != ErrRoomFull {
		t.Fatalf("third Join() error = %v, want ErrRoomFull", err)
	}
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "room.start", Seq: 1})
	if r.phase != PhaseSetup {
		t.Fatalf("phase = %s, want setup", r.phase)
	}
	if r.seats[game.East] != nil || r.seats[game.West] != nil {
		t.Fatalf("junqi room should not allocate side seats")
	}
	r.mu.Lock()
	r.cancelSetupTimerLocked()
	r.mu.Unlock()
}

func TestJunqiConcedeGivesOpponentSingleWinner(t *testing.T) {
	r := newPlayingJunqiRoom(t)
	north := r.seats[game.North]
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "concede", Seq: 1})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended", r.phase)
	}
	if len(r.state.Winner) != 1 || r.state.Winner[0] != game.South {
		t.Fatalf("winner = %v, want South", r.state.Winner)
	}
}

func TestJunqiBlueConcedeGivesRedWinner(t *testing.T) {
	r := newPlayingJunqiRoom(t)
	south := r.seats[game.South]
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "concede", Seq: 1})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended", r.phase)
	}
	if len(r.state.Winner) != 1 || r.state.Winner[0] != game.North {
		t.Fatalf("winner = %v, want North", r.state.Winner)
	}
}

func TestJunqiBlueSurrenderRequestGivesRedWinner(t *testing.T) {
	r := newPlayingJunqiRoom(t)
	south := r.seats[game.South]
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "request.surrender", Seq: 1})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended", r.phase)
	}
	if len(r.state.Winner) != 1 || r.state.Winner[0] != game.North {
		t.Fatalf("winner = %v, want North", r.state.Winner)
	}
}

func TestJunqiBlueFifthSkipGivesRedWinner(t *testing.T) {
	r := newPlayingJunqiRoom(t)
	south := r.seats[game.South]
	r.state.Turn = game.South
	r.skips[game.South] = protocol.MaxSkipsPerPlayer - 1
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "move.skip", Seq: 1})
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended", r.phase)
	}
	if len(r.state.Winner) != 1 || r.state.Winner[0] != game.North {
		t.Fatalf("winner = %v, want North", r.state.Winner)
	}
}

func TestSkipBlockedWhileRequestPending(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "request.tie", Seq: 1})
	if r.request == nil {
		t.Fatalf("expected pending request")
	}
	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "move.skip", Seq: 2})
	if r.skips[game.North] != 0 {
		t.Fatalf("skip should not register while request pending")
	}
}

func TestAutoTimeoutAfterMaxSkipsConcedes(t *testing.T) {
	r := newPlayingRoom(t)
	r.skips[game.North] = protocol.MaxSkipsPerPlayer
	r.mu.Lock()
	r.onTurnTimeoutLocked(game.North)
	r.mu.Unlock()
	if r.phase != PhaseEnded {
		t.Fatalf("phase = %s, want ended after timeout at max skips", r.phase)
	}
	if len(r.state.Winner) != 2 || !(r.state.Winner[0].SameTeam(game.East) && r.state.Winner[1].SameTeam(game.East)) {
		t.Fatalf("winner = %v, want EW team", r.state.Winner)
	}
}

func TestRivalRejectingTieResumesPlay(t *testing.T) {
	r := newPlayingRoom(t)
	north := r.seats[game.North]
	south := r.seats[game.South]
	east := r.seats[game.East]

	sendTestMessage(t, r, north.Token, protocol.ClientMessage{Type: "request.tie", Seq: 1})
	sendTestMessage(t, r, south.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: true})
	sendTestMessage(t, r, east.Token, protocol.ClientMessage{Type: "request.respond", Seq: 1, Accept: false})
	if r.request != nil {
		t.Fatalf("request should be cleared after rival rejection")
	}
	if r.phase != PhasePlaying {
		t.Fatalf("phase = %s, want playing after rejection", r.phase)
	}
}

func newPlayingRoom(t *testing.T) *Room {
	t.Helper()
	r := New("ABC123")
	for _, name := range []string{"north", "east", "south", "west"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	state, err := game.NewState(
		[]game.Piece{
			{ID: 1, Owner: game.North, Rank: game.Commander, Alive: true},
			{ID: 2, Owner: game.East, Rank: game.Commander, Alive: true},
			{ID: 3, Owner: game.South, Rank: game.Commander, Alive: true},
			{ID: 4, Owner: game.West, Rank: game.Commander, Alive: true},
		},
		map[game.PieceID]game.Pos{
			1: {Row: 0, Col: 8},
			2: {Row: 8, Col: 16},
			3: {Row: 16, Col: 8},
			4: {Row: 8, Col: 0},
		},
		game.North,
	)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	r.state = state
	r.phase = PhasePlaying
	r.skips = map[game.Seat]int{}
	r.cancelTurnTimerLocked()
	return r
}

func newPlayingJunqiRoom(t *testing.T) *Room {
	t.Helper()
	r := New("DUEL01")
	r.ConfigureInitial(game.ModeJunqi, nil, nil)
	for _, name := range []string{"north", "south"} {
		if _, err := r.Join(name, ""); err != nil {
			t.Fatalf("Join(%s) error = %v", name, err)
		}
	}
	state, err := game.NewStateForMode(
		game.ModeJunqi,
		[]game.Piece{
			{ID: 1, Owner: game.North, Rank: game.Commander, Alive: true},
			{ID: 2, Owner: game.South, Rank: game.Commander, Alive: true},
		},
		map[game.PieceID]game.Pos{
			1: {Row: 3, Col: 8},
			2: {Row: 12, Col: 8},
		},
		game.North,
	)
	if err != nil {
		t.Fatalf("NewStateForMode() error = %v", err)
	}
	r.state = state
	r.phase = PhasePlaying
	r.skips = map[game.Seat]int{}
	r.cancelTurnTimerLocked()
	return r
}

func sendTestMessage(t *testing.T, r *Room, token string, msg protocol.ClientMessage) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	r.Handle(token, data)
}

func sendViewerTestMessage(t *testing.T, r *Room, viewerID string, msg protocol.ClientMessage) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	r.HandleViewer(viewerID, data)
}

func nextServerMessage(t *testing.T, ch <-chan []byte) protocol.ServerMessage {
	t.Helper()
	select {
	case data := <-ch:
		var msg protocol.ServerMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("Unmarshal(server message) error = %v", err)
		}
		return msg
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for server message")
		return protocol.ServerMessage{}
	}
}

func drainMessages(ch <-chan []byte) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
