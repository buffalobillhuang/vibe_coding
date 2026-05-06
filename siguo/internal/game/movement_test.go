package game

import "testing"

func testState(t *testing.T, pieces []Piece, placements map[PieceID]Pos, turn Seat) *GameState {
	t.Helper()
	g, err := NewState(pieces, placements, turn)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}
	return g
}

func hasMove(moves []Move, to Pos) bool {
	for _, move := range moves {
		if move.To == to {
			return true
		}
	}
	return false
}

func moveTo(moves []Move, to Pos) (Move, bool) {
	for _, move := range moves {
		if move.To == to {
			return move, true
		}
	}
	return Move{}, false
}

func TestImmobilePiecesHaveNoLegalMoves(t *testing.T) {
	for _, rank := range []Rank{Mine, Flag} {
		g := testState(t,
			[]Piece{{ID: 1, Owner: North, Rank: rank, Alive: true}},
			map[PieceID]Pos{1: {1, 8}},
			North,
		)
		if moves := LegalMoves(g, 1); len(moves) != 0 {
			t.Fatalf("%s legal moves = %v, want none", rank, moves)
		}
	}
}

func TestPieceInHQCannotMoveOut(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {0, 7}},
		North,
	)
	if moves := LegalMoves(g, 1); len(moves) != 0 {
		t.Fatalf("legal moves = %v, want none", moves)
	}
}

func TestCannotAttackPieceInCamp(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 7},
			2: {2, 7},
		},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{2, 7}) {
		t.Fatal("expected camp occupant to be protected from attack")
	}
}

func TestCanEnterEmptyCampFromDiagonal(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {1, 6}},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{2, 7}) {
		t.Fatalf("expected piece diagonal to camp to be able to enter")
	}
}

func TestCanLeaveCampToDiagonal(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {2, 7}},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{1, 6}) {
		t.Fatalf("expected piece in camp to be able to leave diagonally")
	}
}

func TestCannotMoveOntoTeammate(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
		},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{2, 8}) {
		t.Fatal("expected teammate-occupied cell to be illegal")
	}
}

func TestStraightRailMovementStopsAtBlocker(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {5, 8},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	if !hasMove(moves, Pos{2, 8}) {
		t.Fatalf("expected straight rail move before center camp, moves = %v", moves)
	}
	if hasMove(moves, Pos{3, 8}) || hasMove(moves, Pos{4, 8}) {
		t.Fatalf("expected center camp to stop rail line, moves = %v", moves)
	}
}

func TestEngineerCanTurnOnRailroad(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {1, 8}},
		North,
	)
	moves := LegalMoves(g, 1)
	if !hasMove(moves, Pos{8, 10}) {
		t.Fatalf("expected engineer to turn through connected railroad, moves = %v", moves)
	}
}

func TestEngineerCanFlyAcrossVisibleRailNetwork(t *testing.T) {
	tests := []struct {
		name string
		from Pos
		to   Pos
	}{
		{"through central vertical track", Pos{1, 8}, Pos{15, 8}},
		{"through central horizontal track", Pos{8, 1}, Pos{8, 15}},
		{"through north west arc", Pos{1, 6}, Pos{6, 4}},
		{"through north east arc", Pos{1, 10}, Pos{6, 12}},
		{"through south west arc", Pos{15, 6}, Pos{10, 4}},
		{"through south east arc", Pos{15, 10}, Pos{10, 12}},
		{"across rail intersections and arcs", Pos{1, 6}, Pos{10, 12}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := testState(t,
				[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
				map[PieceID]Pos{1: tt.from},
				North,
			)
			move, ok := moveTo(LegalMoves(g, 1), tt.to)
			if !ok {
				t.Fatalf("expected engineer to fly from %v to %v", tt.from, tt.to)
			}
			if len(move.Path) < 2 || move.Path[0] != tt.from || move.Path[len(move.Path)-1] != tt.to {
				t.Fatalf("expected rail path from %v to %v, got %v", tt.from, tt.to, move.Path)
			}
		})
	}
}

func TestEngineerRailFlightStopsAtBlocker(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: Commander, Alive: true},
			{ID: 3, Owner: East, Rank: PlatoonLeader, Alive: true},
		},
		map[PieceID]Pos{
			1: {0, 6},
			2: {1, 6},
			3: {2, 6},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	if hasMove(moves, Pos{1, 6}) {
		t.Fatalf("expected teammate blocker to be unenterable")
	}
	if hasMove(moves, Pos{2, 6}) || hasMove(moves, Pos{5, 6}) || hasMove(moves, Pos{6, 6}) {
		t.Fatalf("expected teammate blocker to stop engineer rail flight, moves = %v", moves)
	}
}

func TestEngineerCanAttackFirstEnemyBlockerButCannotPassIt(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: East, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {0, 6},
			2: {1, 6},
			3: {2, 6},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	if !hasMove(moves, Pos{1, 6}) {
		t.Fatalf("expected engineer to be able to attack first enemy blocker")
	}
	if hasMove(moves, Pos{2, 6}) || hasMove(moves, Pos{5, 6}) || hasMove(moves, Pos{6, 6}) {
		t.Fatalf("expected enemy blocker to stop engineer rail flight, moves = %v", moves)
	}
}

func TestThreeThroughTracksConnectOppositeSides(t *testing.T) {
	for _, pos := range []Pos{
		{8, 6}, {8, 8}, {8, 10},
		{6, 8}, {8, 8}, {10, 8},
		{6, 6}, {6, 10}, {10, 6}, {10, 10},
	} {
		if !IsRailroad(pos) {
			t.Fatalf("%v should be railroad", pos)
		}
	}
}

func TestHomeRailTracksMatchSeatOrientation(t *testing.T) {
	for _, pos := range []Pos{
		{1, 6}, {1, 8}, {1, 10}, {5, 6}, {5, 8}, {5, 10},
		{11, 6}, {11, 8}, {11, 10}, {15, 6}, {15, 8}, {15, 10},
		{6, 1}, {8, 1}, {10, 1}, {6, 5}, {8, 5}, {10, 5},
		{6, 11}, {8, 11}, {10, 11}, {6, 15}, {8, 15}, {10, 15},
	} {
		if !IsRailroad(pos) {
			t.Fatalf("%v should be on a visible home rail track", pos)
		}
	}
}

func TestHomeMiddleLinesAreNotRailroads(t *testing.T) {
	for _, pos := range []Pos{
		{2, 8}, {3, 8}, {4, 8},
		{12, 8}, {13, 8}, {14, 8},
		{8, 2}, {8, 3}, {8, 4},
		{8, 12}, {8, 13}, {8, 14},
	} {
		if IsRailroad(pos) {
			t.Fatalf("%v should not be railroad on the visible track map", pos)
		}
	}
}

func TestCentralFrontSegmentsOnlyHaveThreePiecePositions(t *testing.T) {
	for _, pos := range []Pos{{6, 7}, {6, 9}, {7, 6}, {9, 6}, {10, 7}, {10, 9}, {7, 10}, {9, 10}} {
		if IsPlayable(pos) {
			t.Fatalf("%v should be a visual track segment, not a piece position", pos)
		}
	}
	for _, pos := range []Pos{{6, 6}, {6, 8}, {6, 10}, {8, 6}, {8, 8}, {8, 10}, {10, 6}, {10, 8}, {10, 10}} {
		if !IsPlayable(pos) {
			t.Fatalf("%v should be a central track node", pos)
		}
	}
}

func TestAdjacentSideTracksAreConnected(t *testing.T) {
	tests := []struct {
		name string
		from Pos
		to   Pos
	}{
		{"north west connector", Pos{5, 6}, Pos{6, 5}},
		{"west north connector", Pos{6, 5}, Pos{5, 6}},
		{"north east connector", Pos{5, 10}, Pos{6, 11}},
		{"east north connector", Pos{6, 11}, Pos{5, 10}},
		{"south west connector", Pos{11, 6}, Pos{10, 5}},
		{"west south connector", Pos{10, 5}, Pos{11, 6}},
		{"south east connector", Pos{11, 10}, Pos{10, 11}},
		{"east south connector", Pos{10, 11}, Pos{11, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := testState(t,
				[]Piece{
					{ID: 1, Owner: North, Rank: Commander, Alive: true},
					{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
				},
				map[PieceID]Pos{1: tt.from, 2: tt.to},
				North,
			)
			if !hasMove(LegalMoves(g, 1), tt.to) {
				t.Fatalf("expected railway connector from %v to %v", tt.from, tt.to)
			}
		})
	}
}

func TestNormalPieceCanFollowNeighborConnectorFromTrackInterior(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 6},
			2: {6, 4},
		},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{6, 4}) {
		t.Fatalf("expected North side rail to follow connector onto neighbor rail")
	}
}

func TestBombCannotEnterHQ(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Bomb, Alive: true}},
		map[PieceID]Pos{1: {1, 7}},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{0, 7}) {
		t.Fatal("expected bomb to be barred from HQ")
	}
}
