package game

type ClientPiece struct {
	ID      PieceID
	Owner   Seat
	Rank    Rank
	Alive   bool
	Exposed bool
}

type ClientCell struct {
	Pos   Pos
	Type  CellType
	Piece *ClientPiece
}

type ClientView struct {
	Mode       GameMode
	Viewer     Seat
	Phase      Phase
	Turn       Seat
	Cells      []ClientCell
	DeadPieces []ClientPiece
	Eliminated map[Seat]bool
	Winner     []Seat
}

func (g *GameState) ViewFor(viewer Seat) ClientView {
	return g.viewFor(viewer, false)
}

func (g *GameState) ViewForSpectator() ClientView {
	return g.viewFor(North, true)
}

func (g *GameState) viewFor(viewer Seat, spectator bool) ClientView {
	view := ClientView{
		Mode:       g.Mode,
		Viewer:     viewer,
		Phase:      g.Phase,
		Turn:       g.Turn,
		Cells:      make([]ClientCell, 0, BoardSize*BoardSize),
		Eliminated: make(map[Seat]bool, len(g.Eliminated)),
		Winner:     append([]Seat(nil), g.Winner...),
	}
	for seat, eliminated := range g.Eliminated {
		view.Eliminated[seat] = eliminated
	}
	for _, piece := range g.Pieces {
		if piece.Alive {
			continue
		}
		if !spectator && piece.Owner != viewer {
			continue
		}
		view.DeadPieces = append(view.DeadPieces, ClientPiece{
			ID:    piece.ID,
			Owner: piece.Owner,
			Rank:  piece.Rank,
			Alive: false,
		})
	}

	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			pos := Pos{row, col}
			cell := BoardCellForMode(g.Mode, pos)
			if cell.Type == OffBoard {
				continue
			}
			out := ClientCell{Pos: pos, Type: cell.Type}
			if piece, ok := g.PieceAt(pos); ok {
				cp := ClientPiece{
					ID:      piece.ID,
					Owner:   piece.Owner,
					Rank:    Unknown,
					Alive:   piece.Alive,
					Exposed: g.Revealed[piece.ID],
				}
				if (!spectator && piece.Owner == viewer) || g.Revealed[piece.ID] {
					cp.Rank = piece.Rank
				}
				out.Piece = &cp
			}
			view.Cells = append(view.Cells, out)
		}
	}
	return view
}
