package game

import "testing"

func standardJunqiSetup(owner Seat, startID PieceID) ([]Piece, map[PieceID]Pos) {
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
	northPositions := []Pos{
		{2, 7},
		{2, 6}, {2, 8}, {2, 10},
		{2, 9}, {3, 6},
		{3, 7}, {3, 8}, {3, 9},
		{3, 10}, {4, 6}, {4, 8},
		{4, 10}, {5, 6}, {5, 7},
		{5, 9}, {5, 10},
		{6, 6}, {6, 8},
		{6, 10}, {7, 6},
		{7, 7}, {7, 8},
		{7, 9}, {7, 10},
	}
	southPositions := []Pos{
		{14, 7},
		{14, 6}, {14, 8}, {14, 10},
		{14, 9}, {13, 6},
		{13, 7}, {13, 8}, {13, 9},
		{13, 10}, {12, 6}, {12, 8},
		{12, 10}, {11, 6}, {11, 7},
		{11, 9}, {11, 10},
		{10, 6}, {10, 8},
		{10, 10}, {9, 6},
		{9, 7}, {9, 8},
		{9, 9}, {9, 10},
	}
	positions := northPositions
	if owner == South {
		positions = southPositions
	}

	pieces := make([]Piece, len(ranks))
	placements := map[PieceID]Pos{}
	for i, rank := range ranks {
		id := startID + PieceID(i)
		pieces[i] = Piece{ID: id, Owner: owner, Rank: rank, Alive: true}
		placements[id] = positions[i]
	}
	return pieces, placements
}

func TestJunqiValidateSetupAcceptsStandardArmy(t *testing.T) {
	pieces, placements := standardJunqiSetup(North, 1)
	if err := ValidateSetupForMode(ModeJunqi, North, pieces, placements); err != nil {
		t.Fatalf("ValidateSetupForMode() error = %v", err)
	}
}

func TestJunqiTurnsAlternateNorthSouth(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {3, 8}},
		North,
	)
	g.AdvanceTurn()
	if g.Turn != South {
		t.Fatalf("turn = %s, want South", g.Turn)
	}
}

func TestJunqiFlagCaptureWinsOnePlayerGame(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: Flag, Alive: true},
		},
		map[PieceID]Pos{
			1: {13, 7},
			2: {14, 7},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{13, 7}, To: Pos{14, 7}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if next.Phase != Ended {
		t.Fatalf("phase = %v, want Ended", next.Phase)
	}
	if len(next.Winner) != 1 || next.Winner[0] != North {
		t.Fatalf("winner = %v, want North", next.Winner)
	}
}

func TestJunqiMountainGapOnlyAllowsThreeCrossings(t *testing.T) {
	for _, col := range []int{7, 9} {
		if cell := BoardCellForMode(ModeJunqi, Pos{8, col}); cell.Type != OffBoard {
			t.Fatalf("mountain gap at col %d type = %v, want OffBoard", col, cell.Type)
		}
	}
	for _, col := range []int{6, 8, 10} {
		if cell := BoardCellForMode(ModeJunqi, Pos{8, col}); cell.Type == OffBoard {
			t.Fatalf("crossing at col %d is off-board", col)
		}
	}

	blocked := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {7, 7}},
		North,
	)
	if hasMove(LegalMoves(blocked, 1), Pos{8, 7}) {
		t.Fatalf("piece should not cross mountain gap away from the three crossings")
	}

	open := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {7, 8}},
		North,
	)
	if !hasMove(LegalMoves(open, 1), Pos{8, 8}) {
		t.Fatalf("piece should cross through the center mountain crossing")
	}
}

func testStateForMode(t *testing.T, mode GameMode, pieces []Piece, placements map[PieceID]Pos, turn Seat) *GameState {
	t.Helper()
	g, err := NewStateForMode(mode, pieces, placements, turn)
	if err != nil {
		t.Fatalf("NewStateForMode() error = %v", err)
	}
	return g
}
