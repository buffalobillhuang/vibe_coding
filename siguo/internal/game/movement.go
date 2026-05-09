package game

func LegalMoves(g *GameState, pieceID PieceID) []Move {
	piece, ok := g.Pieces[pieceID]
	if !ok || !piece.Alive || !piece.Rank.Movable() || g.Eliminated[piece.Owner] {
		return nil
	}
	from, ok := g.positionOf(pieceID)
	if !ok || g.boardCell(from).Type == HQ {
		return nil
	}

	seen := map[Pos]bool{}
	var moves []Move
	add := func(to Pos, path []Pos) {
		if seen[to] || !canEnter(g, piece, to) {
			return
		}
		seen[to] = true
		moves = append(moves, Move{PieceID: pieceID, From: from, To: to, Path: append([]Pos(nil), path...)})
	}
	addRoad := func(to Pos) {
		if g.boardCell(to).Type == Mountain {
			return
		}
		add(to, nil)
	}

	if g.Mode == ModeJunqi {
		for _, to := range junqiRoadAdjacency(from) {
			addRoad(to)
		}
	} else {
		for _, d := range orthogonalDirs {
			to := Pos{from.Row + d.Row, from.Col + d.Col}
			addRoad(to)
		}
		for _, d := range diagonalDirs {
			to := Pos{from.Row + d.Row, from.Col + d.Col}
			if g.boardCell(from).Type == Camp || g.boardCell(to).Type == Camp {
				addRoad(to)
			}
		}
	}

	if piece.Rank == Engineer {
		if g.isRailroad(from) || g.boardCell(from).Type == Mountain {
			for to, path := range engineerRailMoves(g, piece, from) {
				add(to, path)
			}
		}
		return moves
	}

	if !g.isRailroad(from) {
		return moves
	}

	for _, to := range straightRailMoves(g, piece, from) {
		add(to, []Pos{from, to})
	}
	return moves
}

func IsLegalMove(g *GameState, move Move) bool {
	for _, legal := range LegalMoves(g, move.PieceID) {
		if legal.To == move.To {
			return true
		}
	}
	return false
}

func canEnter(g *GameState, mover Piece, to Pos) bool {
	if !to.InBounds() || !g.isPlayable(to) {
		return false
	}
	cell := g.boardCell(to)
	if g.Mode != ModeJunqi && mover.Rank == Bomb && cell.Type == HQ {
		return false
	}
	if cell.Type == Mountain && mover.Rank != Engineer {
		return false
	}
	target, occupied := g.PieceAt(to)
	if !occupied {
		return true
	}
	if cell.Type == Mountain {
		return mover.Rank == Engineer && target.Rank == Engineer && !g.sameSide(mover.Owner, target.Owner)
	}
	if g.sameSide(mover.Owner, target.Owner) {
		return false
	}
	if cell.Type == Camp {
		return false
	}
	return true
}

func straightRailMoves(g *GameState, mover Piece, from Pos) []Pos {
	out := map[Pos]bool{}
	if g.Mode != ModeJunqi {
		if next, exitDir, ok := directConnectorFrom(from); ok {
			if occupant, occupied := g.PieceAt(next); occupied {
				if !g.sameSide(mover.Owner, occupant.Owner) && g.boardCell(next).Type != Camp {
					out[next] = true
				}
			} else {
				out[next] = true
				for _, dest := range continueRailLine(g, mover, next, exitDir, 1) {
					out[dest] = true
				}
			}
		}
	}
	for _, first := range g.railroadAdj()[from] {
		dir := edgeDir(from, first)
		for _, dest := range continueRailLine(g, mover, from, dir, 0) {
			out[dest] = true
		}
	}
	if g.Mode != ModeJunqi {
		if to, ok := railConnectorPairs()[from]; ok {
			if canEnter(g, mover, to) {
				out[to] = true
			}
		}
	}

	moves := make([]Pos, 0, len(out))
	for pos := range out {
		moves = append(moves, pos)
	}
	return moves
}

func continueRailLine(g *GameState, mover Piece, from Pos, dir Pos, turns int) []Pos {
	type node struct {
		pos   Pos
		dir   Pos
		turns int
	}
	out := map[Pos]bool{}
	queue := []node{{pos: from, dir: dir, turns: turns}}
	visited := map[node]bool{}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true

		for _, next := range g.railroadAdj()[cur.pos] {
			nextDir := edgeDir(cur.pos, next)
			if nextDir != cur.dir {
				continue
			}
			if occupant, occupied := g.PieceAt(next); occupied {
				if !g.sameSide(mover.Owner, occupant.Owner) && g.boardCell(next).Type != Camp {
					out[next] = true
				}
				continue
			}
			out[next] = true
			queue = append(queue, node{pos: next, dir: nextDir, turns: cur.turns})
		}
		if g.Mode != ModeJunqi && cur.turns == 0 {
			if next, exitDir, ok := connectorTransition(cur.pos, cur.dir); ok {
				if occupant, occupied := g.PieceAt(next); occupied {
					if !g.sameSide(mover.Owner, occupant.Owner) && g.boardCell(next).Type != Camp {
						out[next] = true
					}
				} else {
					out[next] = true
					queue = append(queue, node{pos: next, dir: exitDir, turns: cur.turns + 1})
				}
			}
		}
	}

	moves := make([]Pos, 0, len(out))
	for pos := range out {
		moves = append(moves, pos)
	}
	return moves
}

func (g *GameState) boardCell(pos Pos) Cell {
	return BoardCellForMode(g.Mode, pos)
}

func (g *GameState) isPlayable(pos Pos) bool {
	return IsPlayableForMode(g.Mode, pos)
}

func (g *GameState) isRailroad(pos Pos) bool {
	return IsRailroadForMode(g.Mode, pos)
}

func (g *GameState) railroadAdj() map[Pos][]Pos {
	if g.Mode == ModeJunqi {
		return junqiRailroadAdj
	}
	return railroadAdj
}

func (g *GameState) sameSide(a, b Seat) bool {
	if g.Mode == ModeJunqi {
		return a == b
	}
	return a.SameTeam(b)
}

func junqiRoadAdjacency(from Pos) []Pos {
	out := make([]Pos, 0, 8)
	add := func(to Pos) {
		if BoardCellForMode(ModeJunqi, to).Type != OffBoard {
			out = append(out, to)
		}
	}
	for _, d := range orthogonalDirs {
		to := Pos{from.Row + d.Row, from.Col + d.Col}
		add(to)
	}
	for _, to := range junqiCampDiagonalNeighbors(from) {
		add(to)
	}
	return out
}

func junqiCampDiagonalNeighbors(from Pos) []Pos {
	camps := []Pos{{4, 7}, {4, 9}, {5, 8}, {6, 7}, {6, 9}, {10, 7}, {10, 9}, {11, 8}, {12, 7}, {12, 9}}
	out := make([]Pos, 0, 4)
	addAroundCamp := func(camp Pos) {
		for _, d := range diagonalDirs {
			n := Pos{camp.Row + d.Row, camp.Col + d.Col}
			if from == camp && BoardCellForMode(ModeJunqi, n).Type != OffBoard {
				out = append(out, n)
			}
			if from == n {
				out = append(out, camp)
			}
		}
	}
	for _, camp := range camps {
		addAroundCamp(camp)
	}
	return out
}

func edgeDir(from, to Pos) Pos {
	return Pos{sign(to.Row - from.Row), sign(to.Col - from.Col)}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func connectorTransition(pos Pos, dir Pos) (Pos, Pos, bool) {
	switch pos {
	case Pos{6, 6}:
		return Pos{}, Pos{}, false
	case Pos{6, 10}:
		return Pos{}, Pos{}, false
	case Pos{10, 6}:
		return Pos{}, Pos{}, false
	case Pos{10, 10}:
		return Pos{}, Pos{}, false
	default:
		return directConnectorTransition(pos, dir)
	}
}

func directConnectorTransition(pos Pos, dir Pos) (Pos, Pos, bool) {
	type transition struct {
		from Pos
		in   Pos
		to   Pos
		out  Pos
	}
	transitions := []transition{
		{from: Pos{5, 6}, in: Pos{1, 0}, to: Pos{6, 5}, out: Pos{0, -1}},
		{from: Pos{6, 5}, in: Pos{0, 1}, to: Pos{5, 6}, out: Pos{-1, 0}},
		{from: Pos{5, 10}, in: Pos{1, 0}, to: Pos{6, 11}, out: Pos{0, 1}},
		{from: Pos{6, 11}, in: Pos{0, -1}, to: Pos{5, 10}, out: Pos{-1, 0}},
		{from: Pos{11, 6}, in: Pos{-1, 0}, to: Pos{10, 5}, out: Pos{0, -1}},
		{from: Pos{10, 5}, in: Pos{0, 1}, to: Pos{11, 6}, out: Pos{1, 0}},
		{from: Pos{11, 10}, in: Pos{-1, 0}, to: Pos{10, 11}, out: Pos{0, 1}},
		{from: Pos{10, 11}, in: Pos{0, -1}, to: Pos{11, 10}, out: Pos{1, 0}},
	}
	for _, t := range transitions {
		if t.from == pos && t.in == dir {
			return t.to, t.out, true
		}
	}
	return Pos{}, Pos{}, false
}

func directConnectorFrom(pos Pos) (Pos, Pos, bool) {
	type transition struct {
		from Pos
		to   Pos
		out  Pos
	}
	transitions := []transition{
		{from: Pos{5, 6}, to: Pos{6, 5}, out: Pos{0, -1}},
		{from: Pos{6, 5}, to: Pos{5, 6}, out: Pos{-1, 0}},
		{from: Pos{5, 10}, to: Pos{6, 11}, out: Pos{0, 1}},
		{from: Pos{6, 11}, to: Pos{5, 10}, out: Pos{-1, 0}},
		{from: Pos{11, 6}, to: Pos{10, 5}, out: Pos{0, -1}},
		{from: Pos{10, 5}, to: Pos{11, 6}, out: Pos{1, 0}},
		{from: Pos{11, 10}, to: Pos{10, 11}, out: Pos{0, 1}},
		{from: Pos{10, 11}, to: Pos{11, 10}, out: Pos{1, 0}},
	}
	for _, t := range transitions {
		if t.from == pos {
			return t.to, t.out, true
		}
	}
	return Pos{}, Pos{}, false
}

func engineerRailMoves(g *GameState, mover Piece, from Pos) map[Pos][]Pos {
	type node struct {
		pos  Pos
		path []Pos
	}

	out := map[Pos][]Pos{}
	visited := map[Pos]bool{}
	var queue []node

	if g.isRailroad(from) {
		visited[from] = true
		queue = append(queue, node{pos: from, path: []Pos{from}})
		for _, mountain := range adjacentMountainCells(g.Mode, from) {
			if canEnter(g, mover, mountain) {
				out[mountain] = []Pos{from, mountain}
			}
		}
	} else if g.boardCell(from).Type == Mountain {
		for to, path := range adjacentMountainRailMoves(g, mover, from) {
			out[to] = path
		}
		for _, rail := range adjacentRailCells(g.Mode, from) {
			if visited[rail] || !canEnter(g, mover, rail) {
				continue
			}
			visited[rail] = true
			path := []Pos{from, rail}
			out[rail] = path
			if _, occupied := g.PieceAt(rail); !occupied {
				queue = append(queue, node{pos: rail, path: path})
			}
		}
	} else {
		return out
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, mountain := range adjacentMountainCells(g.Mode, cur.pos) {
			if canEnter(g, mover, mountain) {
				if _, seen := out[mountain]; !seen {
					out[mountain] = append(append([]Pos(nil), cur.path...), mountain)
				}
			}
		}

		for _, next := range g.railroadAdj()[cur.pos] {
			if visited[next] || !canEnter(g, mover, next) {
				continue
			}
			visited[next] = true

			path := append(append([]Pos(nil), cur.path...), next)
			if _, occupied := g.PieceAt(next); occupied {
				out[next] = path
				continue
			}

			out[next] = path
			queue = append(queue, node{pos: next, path: path})
		}
	}

	return out
}

func adjacentMountainRailMoves(g *GameState, mover Piece, from Pos) map[Pos][]Pos {
	out := map[Pos][]Pos{}
	for _, to := range mountainCellsForMode(g.Mode) {
		if !connectedMountains(from, to) || !canEnter(g, mover, to) {
			continue
		}
		out[to] = []Pos{from, to}
	}
	return out
}

func connectedMountains(a, b Pos) bool {
	dr := abs(a.Row - b.Row)
	dc := abs(a.Col - b.Col)
	return dr <= 2 && dc <= 2 && (dr != 0 || dc != 0)
}

func junqiAdjacentMountainCells(pos Pos) []Pos {
	seen := map[Pos]bool{}
	for _, mountain := range junqiMountainCells() {
		if adjacent(pos, mountain) {
			seen[mountain] = true
		}
	}
	out := make([]Pos, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

func siguoAdjacentMountainCells(pos Pos) []Pos {
	seen := map[Pos]bool{}
	for _, mountain := range siguoMountainCells() {
		if adjacent(pos, mountain) {
			seen[mountain] = true
		}
	}
	out := make([]Pos, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

func adjacentMountainCells(mode GameMode, pos Pos) []Pos {
	if mode == ModeJunqi {
		return junqiAdjacentMountainCells(pos)
	}
	return siguoAdjacentMountainCells(pos)
}

func mountainCellsForMode(mode GameMode) []Pos {
	if mode == ModeJunqi {
		return junqiMountainCells()
	}
	return siguoMountainCells()
}

func junqiAdjacentRailCells(pos Pos) []Pos {
	return adjacentRailCellsFromMountain(ModeJunqi, pos)
}

func siguoAdjacentRailCells(pos Pos) []Pos {
	return adjacentRailCellsFromMountain(ModeSiguo, pos)
}

func adjacentRailCellsFromMountain(mode GameMode, pos Pos) []Pos {
	if BoardCellForMode(mode, pos).Type != Mountain {
		return nil
	}
	seen := map[Pos]bool{}
	add := func(p Pos) {
		if IsRailroadForMode(mode, p) {
			seen[p] = true
		}
	}
	for _, d := range eightDirs {
		add(Pos{pos.Row + d.Row, pos.Col + d.Col})
	}
	out := make([]Pos, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

func junqiMountainCells() []Pos {
	return []Pos{{8, 7}, {8, 9}}
}

func adjacent(a, b Pos) bool {
	dr := abs(a.Row - b.Row)
	dc := abs(a.Col - b.Col)
	return dr <= 1 && dc <= 1 && (dr != 0 || dc != 0)
}

func adjacentRailCells(mode GameMode, pos Pos) []Pos {
	if mode == ModeJunqi {
		return junqiAdjacentRailCells(pos)
	}
	return siguoAdjacentRailCells(pos)
}
