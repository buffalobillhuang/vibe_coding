package game

import "testing"

func TestApplyMoveAttackerWinsCombatAndAdvancesTurn(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
		},
		North,
	)

	next, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if len(events) != 1 || events[0].Type != EventCombat || events[0].Outcome != AttackerWins {
		t.Fatalf("events = %#v, want attacker combat win", events)
	}
	if !samePath(events[0].Path, []Pos{{1, 8}, {2, 8}}) {
		t.Fatalf("combat event path = %v, want [{1 8} {2 8}]", events[0].Path)
	}
	if _, ok := next.positionOf(2); ok {
		t.Fatal("defender still has a position")
	}
	if pos, _ := next.positionOf(1); pos != (Pos{2, 8}) {
		t.Fatalf("attacker position = %v, want {2 8}", pos)
	}
	if next.Turn != West {
		t.Fatalf("turn = %s, want West", next.Turn)
	}
}

func TestApplyMoveIncludesEngineerRailPathInEvent(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {1, 6}},
		North,
	)

	_, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 6}, To: Pos{10, 12}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if len(events) != 1 || events[0].Type != EventMove {
		t.Fatalf("events = %#v, want move event", events)
	}
	if len(events[0].Path) < 3 || events[0].Path[0] != (Pos{1, 6}) || events[0].Path[len(events[0].Path)-1] != (Pos{10, 12}) {
		t.Fatalf("engineer event path = %v, want multi-hop path from {1 6} to {10 12}", events[0].Path)
	}
}

func TestCommanderDeathRevealsOwnersFlag(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Bomb, Alive: true},
			{ID: 3, Owner: North, Rank: Flag, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
			3: {0, 7},
		},
		North,
	)

	next, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if !next.Revealed[3] {
		t.Fatal("expected flag to be revealed after commander death")
	}
	view := next.ViewFor(North)
	flag := pieceAt(view, Pos{0, 7})
	if flag == nil || flag.Rank != Flag || !flag.Exposed {
		t.Fatalf("own flag view = %#v, want exposed flag", flag)
	}
	opponentFlag := pieceAt(next.ViewFor(East), Pos{0, 7})
	if opponentFlag == nil || opponentFlag.Rank != Flag {
		t.Fatalf("opponent flag view = %#v, want visible flag", opponentFlag)
	}
	if len(events) != 2 || events[1].Type != EventFlagRevealed {
		t.Fatalf("events = %#v, want flag reveal event", events)
	}
}

func TestFlagCaptureEliminatesOwnerAndCanEndGame(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Engineer, Alive: true},
			{ID: 2, Owner: East, Rank: Flag, Alive: true},
			{ID: 3, Owner: West, Rank: Flag, Alive: true},
		},
		map[PieceID]Pos{
			1: {8, 15},
			2: {8, 16},
			3: {9, 0},
		},
		North,
	)
	g.Eliminated[West] = true

	next, events, err := ApplyMove(g, Move{PieceID: 1, From: Pos{8, 15}, To: Pos{8, 16}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if !next.Eliminated[East] {
		t.Fatal("expected East to be eliminated")
	}
	if next.Phase != Ended {
		t.Fatalf("phase = %v, want Ended", next.Phase)
	}
	if len(next.Winner) != 2 || next.Winner[0] != North || next.Winner[1] != South {
		t.Fatalf("winner = %v, want North+South", next.Winner)
	}
	if events[len(events)-2].Type != EventTeamEliminated {
		t.Fatalf("penultimate event = %#v, want team eliminated", events[len(events)-2])
	}
	if events[len(events)-1].Type != EventGameEnded {
		t.Fatalf("last event = %#v, want game ended", events[len(events)-1])
	}
}

func TestApplyMoveSkipsEliminatedPlayerTurn(t *testing.T) {
	g := testState(t,
		[]Piece{{ID: 1, Owner: North, Rank: Engineer, Alive: true}},
		map[PieceID]Pos{1: {1, 8}},
		North,
	)
	g.Eliminated[East] = true

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	if next.Turn != West {
		t.Fatalf("turn = %s, want West", next.Turn)
	}
}

func TestViewForHidesTeammateAndOpponentRanks(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: South, Rank: Flag, Alive: true},
			{ID: 3, Owner: East, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {16, 7},
			3: {8, 16},
		},
		North,
	)
	g.Revealed[3] = true

	view := g.ViewFor(North)
	seen := map[PieceID]Rank{}
	for _, cell := range view.Cells {
		if cell.Piece != nil {
			seen[cell.Piece.ID] = cell.Piece.Rank
		}
	}

	if seen[1] != Commander {
		t.Fatalf("own rank = %s, want Commander", seen[1])
	}
	if seen[2] != Unknown {
		t.Fatalf("teammate rank = %s, want Unknown", seen[2])
	}
	if seen[3] != Engineer {
		t.Fatalf("revealed opponent rank = %s, want Engineer", seen[3])
	}
}

func TestCombatWinnerRemainsHiddenFromOtherPlayers(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}

	ownerView := next.ViewFor(North)
	opponentView := next.ViewFor(East)

	if rankAt(ownerView, Pos{2, 8}) != Commander {
		t.Fatalf("owner sees winner rank = %s, want Commander", rankAt(ownerView, Pos{2, 8}))
	}
	if rankAt(opponentView, Pos{2, 8}) != Unknown {
		t.Fatalf("opponent sees winner rank = %s, want Unknown", rankAt(opponentView, Pos{2, 8}))
	}
}

func TestViewForOnlyIncludesViewerDeadPieces(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	northView := next.ViewFor(North)
	eastView := next.ViewFor(East)
	if len(northView.DeadPieces) != 0 {
		t.Fatalf("north dead pieces = %v, want none", northView.DeadPieces)
	}
	if len(eastView.DeadPieces) != 1 || eastView.DeadPieces[0].Rank != Engineer {
		t.Fatalf("east dead pieces = %v, want own dead engineer", eastView.DeadPieces)
	}
}

func TestViewForSpectatorIncludesDeadPiecesDuringActivePlay(t *testing.T) {
	g := testState(t,
		[]Piece{
			{ID: 1, Owner: North, Rank: Commander, Alive: true},
			{ID: 2, Owner: East, Rank: Engineer, Alive: true},
			{ID: 3, Owner: West, Rank: Commander, Alive: true},
			{ID: 4, Owner: South, Rank: Engineer, Alive: true},
		},
		map[PieceID]Pos{
			1: {1, 8},
			2: {2, 8},
			3: {8, 1},
			4: {8, 2},
		},
		North,
	)

	next, _, err := ApplyMove(g, Move{PieceID: 1, From: Pos{1, 8}, To: Pos{2, 8}})
	if err != nil {
		t.Fatalf("ApplyMove() error = %v", err)
	}
	next, _, err = ApplyMove(next, Move{PieceID: 3, From: Pos{8, 1}, To: Pos{8, 2}})
	if err != nil {
		t.Fatalf("ApplyMove() second combat error = %v", err)
	}
	view := next.ViewForSpectator()
	if len(view.DeadPieces) != 2 {
		t.Fatalf("spectator dead pieces during active play = %v, want two", view.DeadPieces)
	}
	seen := map[Seat]Rank{}
	for _, dead := range view.DeadPieces {
		seen[dead.Owner] = dead.Rank
	}
	if seen[East] != Engineer || seen[South] != Engineer {
		t.Fatalf("spectator dead pieces during active play = %+v, want east and south engineers", view.DeadPieces)
	}

	next.Phase = Ended
	view = next.ViewForSpectator()
	if len(view.DeadPieces) != 2 {
		t.Fatalf("spectator dead pieces after end = %v, want two", view.DeadPieces)
	}
}

func rankAt(view ClientView, pos Pos) Rank {
	for _, cell := range view.Cells {
		if cell.Pos == pos && cell.Piece != nil {
			return cell.Piece.Rank
		}
	}
	return Unknown
}

func pieceAt(view ClientView, pos Pos) *ClientPiece {
	for _, cell := range view.Cells {
		if cell.Pos == pos {
			return cell.Piece
		}
	}
	return nil
}

func samePath(got, want []Pos) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
