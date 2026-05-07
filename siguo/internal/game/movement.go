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

	for _, d := range orthogonalDirs {
		to := Pos{from.Row + d.Row, from.Col + d.Col}
		add(to, nil)
	}
	for _, d := range diagonalDirs {
		to := Pos{from.Row + d.Row, from.Col + d.Col}
		if g.boardCell(from).Type == Camp || g.boardCell(to).Type == Camp {
			add(to, nil)
		}
	}

	if !g.isRailroad(from) {
		return moves
	}

	if piece.Rank == Engineer {
		for to, path := range engineerRailMoves(g, piece, from) {
			add(to, path)
		}
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
	if mover.Rank == Bomb && cell.Type == HQ {
		return false
	}
	target, occupied := g.PieceAt(to)
	if !occupied {
		return true
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
	visited := map[Pos]bool{from: true}
	queue := []node{{pos: from, path: []Pos{from}}}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, next := range railroadAdj[cur.pos] {
			if visited[next] {
				continue
			}
			visited[next] = true

			path := append(append([]Pos(nil), cur.path...), next)
			if occupant, occupied := g.PieceAt(next); occupied {
				if !mover.Owner.SameTeam(occupant.Owner) && BoardCell(next).Type != Camp {
					out[next] = path
				}
				continue
			}

			out[next] = path
			queue = append(queue, node{pos: next, path: path})
		}
	}

	return out
}
