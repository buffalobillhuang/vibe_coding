package game

import (
	"encoding/json"
	"testing"
)

func TestSeatSlicesMarshalAsJSONNumbers(t *testing.T) {
	data, err := json.Marshal(struct {
		Winner []Seat `json:"winner"`
	}{Winner: []Seat{North, South}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(data) != `{"winner":[0,2]}` {
		t.Fatalf("json = %s, want numeric winner seats", data)
	}
}

func TestSpectatorViewHidesUnrevealedRanks(t *testing.T) {
	state, err := NewState(
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {Row: 1, Col: 8},
			2: {Row: 15, Col: 8},
		},
		North,
	)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	state.Revealed[2] = true

	view := state.ViewForSpectator()
	for _, cell := range view.Cells {
		if cell.Piece == nil {
			continue
		}
		switch cell.Piece.ID {
		case 1:
			if cell.Piece.Rank != Unknown {
				t.Fatalf("unrevealed spectator piece rank = %v, want Unknown", cell.Piece.Rank)
			}
		case 2:
			if cell.Piece.Rank != Engineer {
				t.Fatalf("revealed spectator piece rank = %v, want Engineer", cell.Piece.Rank)
			}
		}
	}
}
