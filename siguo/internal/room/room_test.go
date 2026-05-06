package room

import (
	"encoding/json"
	"testing"

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
	}
}

func sendTestMessage(t *testing.T, r *Room, token string, msg protocol.ClientMessage) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	r.Handle(token, data)
}
