package game

import "errors"

var (
	ErrInvalidPiece   = errors.New("invalid piece")
	ErrPieceNotAlive  = errors.New("piece is not alive")
	ErrPieceImmobile  = errors.New("piece cannot move")
	ErrNotPieceTurn   = errors.New("not piece owner's turn")
	ErrInvalidMove    = errors.New("invalid move")
	ErrDestination    = errors.New("invalid destination")
	ErrGameNotPlaying = errors.New("game is not playing")
)

func NewState(pieces []Piece, placements map[PieceID]Pos, turn Seat) (*GameState, error) {
	g := &GameState{
		Phase:      Playing,
		Turn:       turn,
		Pieces:     make(map[PieceID]Piece, len(pieces)),
		Positions:  make(map[PieceID]Pos, len(pieces)),
		Clocks:     map[Seat]Clock{},
		Revealed:   map[PieceID]bool{},
		Eliminated: map[Seat]bool{},
	}

	for _, piece := range pieces {
		if piece.ID == 0 {
			return nil, ErrInvalidPiece
		}
		if !piece.Alive {
			piece.Alive = true
		}
		pos, ok := placements[piece.ID]
		if !ok || !IsPlayable(pos) {
			return nil, ErrDestination
		}
		if g.BoardIdx[pos.Row][pos.Col] != 0 {
			return nil, ErrDestination
		}
		g.Pieces[piece.ID] = piece
		g.Positions[piece.ID] = pos
		g.BoardIdx[pos.Row][pos.Col] = piece.ID
	}

	return g, nil
}

func (g *GameState) Clone() *GameState {
	clone := *g
	clone.Pieces = make(map[PieceID]Piece, len(g.Pieces))
	for id, piece := range g.Pieces {
		clone.Pieces[id] = piece
	}
	clone.Positions = make(map[PieceID]Pos, len(g.Positions))
	for id, pos := range g.Positions {
		clone.Positions[id] = pos
	}
	clone.Clocks = make(map[Seat]Clock, len(g.Clocks))
	for seat, clock := range g.Clocks {
		clone.Clocks[seat] = clock
	}
	clone.Revealed = make(map[PieceID]bool, len(g.Revealed))
	for id, revealed := range g.Revealed {
		clone.Revealed[id] = revealed
	}
	clone.Eliminated = make(map[Seat]bool, len(g.Eliminated))
	for seat, eliminated := range g.Eliminated {
		clone.Eliminated[seat] = eliminated
	}
	clone.Winner = append([]Seat(nil), g.Winner...)
	clone.History = append([]Event(nil), g.History...)
	return &clone
}

func (g *GameState) PieceAt(pos Pos) (Piece, bool) {
	if !pos.InBounds() {
		return Piece{}, false
	}
	id := g.BoardIdx[pos.Row][pos.Col]
	if id == 0 {
		return Piece{}, false
	}
	piece, ok := g.Pieces[id]
	return piece, ok && piece.Alive
}

func (g *GameState) positionOf(id PieceID) (Pos, bool) {
	pos, ok := g.Positions[id]
	if !ok {
		return Pos{}, false
	}
	piece, ok := g.Pieces[id]
	return pos, ok && piece.Alive
}

func (g *GameState) advanceTurn() {
	next := g.Turn.Next()
	for i := 0; i < 4; i++ {
		if !g.Eliminated[next] {
			g.Turn = next
			return
		}
		next = next.Next()
	}
}

func (g *GameState) kill(id PieceID) {
	piece := g.Pieces[id]
	piece.Alive = false
	g.Pieces[id] = piece
	g.Revealed[id] = true
	if pos, ok := g.Positions[id]; ok {
		g.BoardIdx[pos.Row][pos.Col] = 0
	}
	delete(g.Positions, id)
}

func (g *GameState) movePiece(id PieceID, to Pos) {
	from := g.Positions[id]
	g.BoardIdx[from.Row][from.Col] = 0
	g.BoardIdx[to.Row][to.Col] = id
	g.Positions[id] = to
}

func (g *GameState) revealFlag(owner Seat) (bool, PieceID) {
	for id, piece := range g.Pieces {
		if piece.Owner == owner && piece.Rank == Flag && piece.Alive {
			if !g.Revealed[id] {
				g.Revealed[id] = true
				return true, id
			}
			return false, id
		}
	}
	return false, 0
}

func (g *GameState) eliminate(owner Seat) {
	g.Eliminated[owner] = true
	for id, piece := range g.Pieces {
		if piece.Owner == owner && piece.Alive {
			g.kill(id)
		}
	}
}

func (g *GameState) checkWinner() []Seat {
	nsLost := g.Eliminated[North] && g.Eliminated[South]
	ewLost := g.Eliminated[East] && g.Eliminated[West]
	switch {
	case nsLost && ewLost:
		return []Seat{}
	case nsLost:
		return []Seat{East, West}
	case ewLost:
		return []Seat{North, South}
	default:
		return nil
	}
}
