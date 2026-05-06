package game

import "testing"

func standardNorthSetup() ([]Piece, map[PieceID]Pos) {
	ranks := []Rank{
		Flag,
		Mine, Mine, Mine,
		Bomb, Bomb,
		Engineer, Engineer, Engineer,
		PlatoonLeader, PlatoonLeader, PlatoonLeader,
		CompanyCommander, CompanyCommander, CompanyCommander,
		BattalionCommander, BattalionCommander,
		RegimentCommander, RegimentCommander,
		BrigadeCommander, BrigadeCommander,
		DivisionCommander, DivisionCommander,
		CorpsCommander,
		Commander,
	}
	positions := []Pos{
		{0, 7},
		{0, 6}, {0, 8}, {0, 10},
		{0, 9}, {1, 6},
		{1, 7}, {1, 8}, {1, 9},
		{1, 10}, {2, 6}, {2, 8},
		{2, 10}, {3, 6}, {3, 7},
		{3, 9}, {3, 10},
		{4, 6}, {4, 8},
		{4, 10}, {5, 6},
		{5, 7}, {5, 8},
		{5, 9}, {5, 10},
	}

	pieces := make([]Piece, len(ranks))
	placements := map[PieceID]Pos{}
	for i, rank := range ranks {
		id := PieceID(i + 1)
		pieces[i] = Piece{ID: id, Owner: North, Rank: rank, Alive: true}
		placements[id] = positions[i]
	}
	return pieces, placements
}

func TestValidateSetupAcceptsStandardArmy(t *testing.T) {
	pieces, placements := standardNorthSetup()
	if err := ValidateSetup(North, pieces, placements); err != nil {
		t.Fatalf("ValidateSetup() error = %v", err)
	}
}

func TestValidateSetupRejectsFlagOutsideHQ(t *testing.T) {
	pieces, placements := standardNorthSetup()
	placements[1] = Pos{1, 6}
	if err := ValidateSetup(North, pieces, placements); err == nil {
		t.Fatal("ValidateSetup() error = nil, want flag placement error")
	}
}

func TestValidateSetupAllowsOtherHQOccupied(t *testing.T) {
	pieces, placements := standardNorthSetup()
	if got := placements[5]; got != (Pos{0, 9}) {
		t.Fatalf("expected a non-flag piece in the second HQ, got %v", got)
	}
	if err := ValidateSetup(North, pieces, placements); err != nil {
		t.Fatalf("ValidateSetup() error = %v", err)
	}
}

func TestBoardHasFiveCampsPerHomeZone(t *testing.T) {
	tests := map[Seat]Pos{
		North: {3, 8},
		South: {13, 8},
		West:  {8, 3},
		East:  {8, 13},
	}
	for seat, pos := range tests {
		cell := BoardCell(pos)
		if cell.Type != Camp {
			t.Fatalf("%s center camp at %v type = %v, want Camp", seat, pos, cell.Type)
		}
	}
}

func TestValidateSetupRejectsMineOutsideBackRows(t *testing.T) {
	pieces, placements := standardNorthSetup()
	placements[2] = Pos{3, 6}
	if err := ValidateSetup(North, pieces, placements); err == nil {
		t.Fatal("ValidateSetup() error = nil, want mine placement error")
	}
}

func TestValidateSetupRejectsBombOnFrontRow(t *testing.T) {
	pieces, placements := standardNorthSetup()
	placements[5] = Pos{5, 6}
	if err := ValidateSetup(North, pieces, placements); err == nil {
		t.Fatal("ValidateSetup() error = nil, want bomb placement error")
	}
}
