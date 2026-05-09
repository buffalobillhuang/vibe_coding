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

func TestJunqiValidateSetupRejectsInvalidSpecialPlacements(t *testing.T) {
	t.Run("flag outside HQ", func(t *testing.T) {
		pieces, placements := standardJunqiSetup(North, 1)
		placements[1] = Pos{3, 8}
		if err := ValidateSetupForMode(ModeJunqi, North, pieces, placements); err == nil {
			t.Fatalf("ValidateSetupForMode() error = nil, want flag placement error")
		}
	})
	t.Run("mine outside back rows", func(t *testing.T) {
		pieces, placements := standardJunqiSetup(North, 1)
		placements[2] = Pos{4, 6}
		if err := ValidateSetupForMode(ModeJunqi, North, pieces, placements); err == nil {
			t.Fatalf("ValidateSetupForMode() error = nil, want mine placement error")
		}
	})
	t.Run("bomb on front row", func(t *testing.T) {
		pieces, placements := standardJunqiSetup(North, 1)
		placements[5] = Pos{7, 6}
		if err := ValidateSetupForMode(ModeJunqi, North, pieces, placements); err == nil {
			t.Fatalf("ValidateSetupForMode() error = nil, want bomb placement error")
		}
	})
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

func TestJunqiBombCanCaptureFlag(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Bomb, Alive: true},
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
	if next.Phase != Ended || len(next.Winner) != 1 || next.Winner[0] != North {
		t.Fatalf("phase/winner = %v/%v, want ended North win", next.Phase, next.Winner)
	}
}

func TestJunqiMountainGapOnlyAllowsThreeCrossings(t *testing.T) {
	for _, col := range []int{7, 9} {
		if cell := BoardCellForMode(ModeJunqi, Pos{8, col}); cell.Type != Mountain {
			t.Fatalf("mountain gap at col %d type = %v, want Mountain", col, cell.Type)
		}
	}
	for _, col := range []int{6, 8, 10} {
		cell := BoardCellForMode(ModeJunqi, Pos{8, col})
		if cell.Type != Frontline {
			t.Fatalf("mountain crossing at col %d type = %v, want Frontline", col, cell.Type)
		}
	}

	blocked := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {7, 7}},
		North,
	)
	if hasMove(LegalMoves(blocked, 1), Pos{8, 7}) {
		t.Fatalf("non-engineer should not enter mountain gap away from the three crossings")
	}

	for _, col := range []int{6, 8, 10} {
		open := testStateForMode(t, ModeJunqi,
			[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
			map[PieceID]Pos{1: {7, col}},
			North,
		)
		if !hasMove(LegalMoves(open, 1), Pos{8, col}) {
			t.Fatalf("piece should enter mountain crossing at col %d", col)
		}
		if !hasMove(LegalMoves(open, 1), Pos{9, col}) {
			t.Fatalf("rail piece should cross through empty mountain crossing at col %d", col)
		}
	}
}

func TestJunqiEngineerCanRailMoveToMountainCells(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {7, 8}},
		North,
	)
	moves := LegalMoves(g, 1)
	for _, mountain := range []Pos{{8, 7}, {8, 9}} {
		if !hasMove(moves, mountain) {
			t.Fatalf("engineer should be able to rail move onto mountain cell %v", mountain)
		}
	}
}

func TestJunqiEngineerCanEnterMountainFromEveryAdjacentRailStation(t *testing.T) {
	tests := []struct {
		from Pos
		to   Pos
	}{
		{Pos{7, 8}, Pos{8, 9}},
		{Pos{7, 10}, Pos{8, 9}},
		{Pos{8, 8}, Pos{8, 9}},
		{Pos{8, 10}, Pos{8, 9}},
		{Pos{9, 8}, Pos{8, 9}},
		{Pos{9, 10}, Pos{8, 9}},
	}
	for _, tt := range tests {
		g := testStateForMode(t, ModeJunqi,
			[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
			map[PieceID]Pos{1: tt.from},
			North,
		)
		if !hasMove(LegalMoves(g, 1), tt.to) {
			t.Fatalf("engineer at adjacent rail %v should enter mountain %v", tt.from, tt.to)
		}
	}
}

func TestJunqiEngineerCanRailFlyThenEnterMountain(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {3, 10}},
		North,
	)
	move, ok := moveTo(LegalMoves(g, 1), Pos{8, 9})
	if !ok {
		t.Fatalf("engineer should rail-fly to a mountain-side track and enter mountain")
	}
	if len(move.Path) < 3 || move.Path[0] != (Pos{3, 10}) || move.Path[len(move.Path)-1] != (Pos{8, 9}) {
		t.Fatalf("expected rail path into mountain, got %v", move.Path)
	}
}

func TestJunqiEngineerOnMountainCanAirDropToEmptyRail(t *testing.T) {
	tests := []struct {
		name string
		from Pos
		to   Pos
	}{
		{"left mountain to south east rail", Pos{8, 7}, Pos{13, 10}},
		{"right mountain to north west rail", Pos{8, 9}, Pos{3, 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := testStateForMode(t, ModeJunqi,
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

func TestJunqiEngineerCanFlyBetweenAdjacentMountains(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {8, 7}},
		North,
	)
	move, ok := moveTo(LegalMoves(g, 1), Pos{8, 9})
	if !ok {
		t.Fatalf("expected engineer to fly between adjacent mountain gaps")
	}
	if len(move.Path) != 2 || move.Path[0] != (Pos{8, 7}) || move.Path[1] != (Pos{8, 9}) {
		t.Fatalf("expected mountain tunnel path, got %v", move.Path)
	}
}

func TestJunqiEngineerCanFlyBetweenAdjacentMountainsThroughTunnelWithBlockedRail(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: South, Rank: CompanyCommander, Alive: true},
		},
		map[PieceID]Pos{
			1: {8, 7},
			2: {7, 8},
			3: {9, 8},
		},
		North,
	)
	if !hasMove(LegalMoves(g, 1), Pos{8, 9}) {
		t.Fatalf("engineer should fly between mountain gaps through the tunnel even when rail points are blocked")
	}
}

func TestJunqiEngineerOnMountainCannotLandOnOccupiedRail(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: PlatoonLeader, Alive: true},
			{ID: 3, Owner: South, Rank: CompanyCommander, Alive: true},
			{ID: 4, Owner: South, Rank: BattalionCommander, Alive: true},
			{ID: 5, Owner: South, Rank: RegimentCommander, Alive: true},
			{ID: 6, Owner: South, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {8, 7},
			2: {7, 6},
			3: {7, 8},
			4: {9, 6},
			5: {9, 8},
			6: {6, 6},
		},
		North,
	)
	moves := LegalMoves(g, 1)
	for _, blocker := range []Pos{{7, 6}, {7, 8}, {9, 6}, {9, 8}} {
		if hasMove(moves, blocker) {
			t.Fatalf("engineer should not attack occupied rail when leaving mountain: %v", blocker)
		}
	}
	if !hasMove(moves, Pos{13, 10}) {
		t.Fatalf("engineer should still air-drop to any empty rail station")
	}
}

func TestJunqiEngineerCanAttackEnemyEngineerInMountain(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
			{ID: 3, Owner: North, Rank: Commander, Alive: true},
			{ID: 4, Owner: South, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 8},
			2: {8, 7},
			3: {3, 6},
			4: {13, 6},
		},
		North,
	)

	if !hasMove(LegalMoves(g, 1), Pos{8, 7}) {
		t.Fatalf("engineer should be able to attack enemy engineer in mountain")
	}
	next, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{7, 8}, To: Pos{8, 7}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if len(events) == 0 || events[0].Type != EventCombat || events[0].Outcome != BothDie {
		t.Fatalf("events = %#v, want both-die engineer combat", events)
	}
	if next.Pieces[1].Alive || next.Pieces[2].Alive {
		t.Fatalf("engineers alive = %v/%v, want both dead", next.Pieces[1].Alive, next.Pieces[2].Alive)
	}
}

func TestJunqiNonEngineerCannotAttackEngineerInMountain(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Bomb, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 8},
			2: {8, 7},
		},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{8, 7}) {
		t.Fatalf("non-engineer should not attack engineer in mountain")
	}
	if _, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{7, 8}, To: Pos{8, 7}}); err == nil {
		t.Fatalf("non-engineer attacking engineer in mountain should be illegal")
	}
}

func TestJunqiEngineerRevealsWhenEnteringMountainCell(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {7, 8},
			2: {13, 6},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{7, 8}, To: Pos{8, 7}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if !next.Revealed[1] {
		t.Fatalf("engineer should be revealed after entering mountain cell")
	}
	if rankAt(next.ViewFor(South), Pos{8, 7}) != Engineer {
		t.Fatalf("opponent should see mountain engineer as revealed")
	}
}

func TestJunqiFrontRowCanAttackAcrossThreeCrossings(t *testing.T) {
	for _, col := range []int{6, 8, 10} {
		g := testStateForMode(t, ModeJunqi,
			[]Piece{
				{ID: 1, Owner: North, Rank: Commander, Alive: true},
				{ID: 2, Owner: South, Rank: Engineer, Alive: true},
				{ID: 3, Owner: South, Rank: Commander, Alive: true},
			},
			map[PieceID]Pos{
				1: {7, col},
				2: {9, col},
				3: {13, 6},
			},
			North,
		)
		next, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{7, col}, To: Pos{9, col}})
		if err != nil {
			t.Fatalf("ApplyMove(crossing col %d) error = %v", col, err)
		}
		if len(events) != 1 || events[0].Type != EventCombat || events[0].Outcome != AttackerWins {
			t.Fatalf("events = %#v, want attacker combat win", events)
		}
		if pos, _ := next.positionOf(1); pos != (Pos{9, col}) {
			t.Fatalf("attacker position = %v, want {9 %d}", pos, col)
		}
	}
}

func TestJunqiMiddleColumnRoadConnectsBackCenterStations(t *testing.T) {
	for _, tt := range []struct {
		name string
		from Pos
		to   Pos
	}{
		{"north center", Pos{3, 8}, Pos{4, 8}},
		{"south center", Pos{12, 8}, Pos{13, 8}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := testStateForMode(t, ModeJunqi,
				[]Piece{{ID: 1, Owner: North, Rank: Bomb, Alive: true}},
				map[PieceID]Pos{1: tt.from},
				North,
			)
			if !hasMove(LegalMoves(g, 1), tt.to) {
				t.Fatalf("expected center road move from %v to %v", tt.from, tt.to)
			}
		})
	}
}

func TestJunqiMiddleRailDoesNotRunIntoCentralCamp(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {4, 8}},
		North,
	)
	moves := LegalMoves(g, 1)
	if hasMove(moves, Pos{7, 8}) {
		t.Fatalf("middle rail should not run through the central camp lattice")
	}
}

func TestJunqiNoMovablePiecesLoses(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: South, Rank: Flag, Alive: true},
			{ID: 3, Owner: South, Rank: Mine, Alive: true},
		},
		map[PieceID]Pos{
			1: {3, 6},
			2: {14, 7},
			3: {14, 6},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{3, 6}, To: Pos{4, 6}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if next.Phase != Ended || len(next.Winner) != 1 || next.Winner[0] != North {
		t.Fatalf("phase/winner = %v/%v, want ended North win", next.Phase, next.Winner)
	}
}

func TestJunqiMoverWithNoMovablePiecesLoses(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Mine, Alive: true},
			{ID: 3, Owner: South, Rank: Commander, Alive: true},
		},
		map[PieceID]Pos{
			1: {13, 7},
			2: {14, 7},
			3: {13, 6},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{13, 7}, To: Pos{14, 7}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if next.Phase != Ended || len(next.Winner) != 1 || next.Winner[0] != South {
		t.Fatalf("phase/winner = %v/%v, want ended South win", next.Phase, next.Winner)
	}
}

func TestJunqiCampOccupantCannotBeAttacked(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {5, 7},
			2: {4, 7},
		},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{4, 7}) {
		t.Fatalf("camp occupant should be protected from attack")
	}
}

func TestJunqiHQPieceCannotMove(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {2, 7}},
		North,
	)
	if moves := LegalMoves(g, 1); len(moves) != 0 {
		t.Fatalf("HQ piece moves = %v, want none", moves)
	}
}

func TestJunqiRoadDiagonalsFollowCampLattice(t *testing.T) {
	g := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {4, 8}},
		North,
	)
	if hasMove(LegalMoves(g, 1), Pos{5, 9}) {
		t.Fatalf("non-camp station should not move diagonally merely because a camp is nearby")
	}

	camp := testStateForMode(t, ModeJunqi,
		[]Piece{{ID: 1, Owner: North, Rank: Commander, Alive: true}},
		map[PieceID]Pos{1: {4, 7}},
		North,
	)
	if !hasMove(LegalMoves(camp, 1), Pos{5, 8}) {
		t.Fatalf("camp should have diagonal road to adjacent mapped station/camp")
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
