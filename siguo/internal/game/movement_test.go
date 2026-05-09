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

func TestSiguoEngineerCanRailFlyThenEnterCenterMountainCells(t *testing.T) {
	for _, pos := range []Pos{{7, 7}, {7, 9}, {9, 7}, {9, 9}} {
		if cell := BoardCell(pos); cell.Type != Mountain {
			t.Fatalf("%v type = %v, want Mountain", pos, cell.Type)
		}
	}

	near := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {6, 6}},
		North,
	)
	move, ok := moveTo(LegalMoves(near, 1), Pos{7, 7})
	if !ok {
		t.Fatalf("expected engineer on adjacent rail station to enter center mountain")
	}
	if len(move.Path) != 2 || move.Path[0] != (Pos{6, 6}) || move.Path[1] != (Pos{7, 7}) {
		t.Fatalf("expected one-step mountain entry, got %v", move.Path)
	}

	far := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {1, 6}},
		North,
	)
	move, ok = moveTo(LegalMoves(far, 1), Pos{7, 7})
	if !ok {
		t.Fatalf("engineer should rail-fly to a mountain-side track and enter mountain")
	}
	if len(move.Path) < 3 || move.Path[0] != (Pos{1, 6}) || move.Path[len(move.Path)-1] != (Pos{7, 7}) {
		t.Fatalf("expected rail path into mountain, got %v", move.Path)
	}
}

func TestSiguoNonEngineerCannotEnterCenterMountain(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {6, 6}},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{7, 7}) {
		t.Fatalf("non-engineer should not enter center mountain")
	}
}

func TestSiguoEngineerCanFlyFromMountainBackToRail(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {7, 7}},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{15, 8}) {
		t.Fatalf("engineer on center mountain should be able to rejoin connected rail network")
	}
}

func TestSiguoEngineerOnMountainCanAirDropToAnyEmptyRail(t *testing.T) {
	tests := []struct {
		name string
		from Pos
		to   Pos
	}{
		{"north west mountain to south home rail", Pos{7, 7}, Pos{15, 8}},
		{"north east mountain to west home rail", Pos{7, 9}, Pos{8, 1}},
		{"south west mountain to east home rail", Pos{9, 7}, Pos{8, 15}},
		{"south east mountain to north home rail", Pos{9, 9}, Pos{1, 8}},
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
				t.Fatalf("expected engineer on mountain %v to fly to rail %v", tt.from, tt.to)
			}
			if len(move.Path) != 2 || move.Path[0] != tt.from || move.Path[1] != tt.to {
				t.Fatalf("expected one-step air-drop from mountain %v to rail %v, got %v", tt.from, tt.to, move.Path)
			}
		})
	}
}

func TestSiguoEngineerCanFlyBetweenMountainCells(t *testing.T) {
	tests := []struct {
		name string
		from Pos
		to   Pos
	}{
		{"top pair", Pos{7, 7}, Pos{7, 9}},
		{"bottom pair", Pos{9, 7}, Pos{9, 9}},
		{"left pair", Pos{7, 7}, Pos{9, 7}},
		{"right pair", Pos{7, 9}, Pos{9, 9}},
		{"down right diagonal", Pos{7, 7}, Pos{9, 9}},
		{"down left diagonal", Pos{7, 9}, Pos{9, 7}},
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
				t.Fatalf("expected engineer to fly between mountains from %v to %v", tt.from, tt.to)
			}
			if len(move.Path) != 2 || move.Path[0] != tt.from || move.Path[1] != tt.to {
				t.Fatalf("expected mountain tunnel path from %v to %v, got %v", tt.from, tt.to, move.Path)
			}
		})
	}
}

func TestSiguoEngineerCanFlyBetweenMountainsThroughTunnelWithBlockedRail(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 7},
			2: {8, 8},
		},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{9, 9}) {
		t.Fatalf("engineer should fly diagonally between mountains through the tunnel even when the rail point is blocked")
	}
}

func TestSiguoEngineerCanFlyBetweenAdjacentMountainsThroughTunnelWithBlockedRail(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: East, Rank: CompanyCommander, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 7},
			2: {6, 8},
			3: {8, 8},
		},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{7, 9}) {
		t.Fatalf("engineer should fly between adjacent mountains through the tunnel even when rail points are blocked")
	}
}

func TestSiguoEngineerOnMountainCannotLandOnOccupiedRail(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: East, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: East, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 7},
			2: {1, 6},
			3: {0, 6},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	if hasMove(moves, Pos{1, 6}) {
		t.Fatalf("engineer should not attack occupied rail when leaving mountain")
	}
	if !hasMove(moves, Pos{5, 6}) {
		t.Fatalf("engineer should still air-drop to any empty rail station")
	}
}

func TestSiguoEngineerOnMountainWithAllSideRailsBlockedCanStillAirDropAndTunnel(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: South, Rank: CompanyCommander, Alive: true},
			{ID: 4, Owner: South, Rank: BattalionCommander, Alive: true},
			{ID: 5, Owner: South, Rank: RegimentCommander, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 7},
			2: {6, 6},
			3: {6, 8},
			4: {8, 6},
			5: {8, 8},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	if !hasMove(moves, Pos{15, 8}) {
		t.Fatalf("engineer should air-drop to empty rail even when adjacent side rails are blocked, moves = %v", moves)
	}
	for _, occupied := range []Pos{{6, 6}, {6, 8}, {8, 6}, {8, 8}} {
		if hasMove(moves, occupied) {
			t.Fatalf("engineer should not land on occupied side rail %v", occupied)
		}
	}
	if !hasMove(moves, Pos{7, 9}) || !hasMove(moves, Pos{9, 7}) || !hasMove(moves, Pos{9, 9}) {
		t.Fatalf("engineer should still use mountain tunnels when rail sides are blocked, moves = %v", moves)
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
