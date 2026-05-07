package game

type CellType uint8

const (
	OffBoard CellType = iota
	Normal
	Camp
	Railroad
	HQ
	Frontline
)

type Cell struct {
	Type CellType
	Home *Seat
}

var defaultBoard [BoardSize][BoardSize]Cell
var junqiBoard [BoardSize][BoardSize]Cell
var railroadAdj map[Pos][]Pos
var junqiRailroadAdj map[Pos][]Pos

func init() {
	defaultBoard = buildBoard()
	junqiBoard = buildJunqiBoard()
	railroadAdj = buildRailroadAdjacency(defaultBoard)
	junqiRailroadAdj = buildOrthogonalRailroadAdjacency(junqiBoard)
}

func BoardCell(pos Pos) Cell {
	return BoardCellForMode(ModeSiguo, pos)
}

func BoardCellForMode(mode GameMode, pos Pos) Cell {
	if !pos.InBounds() {
		return Cell{Type: OffBoard}
	}
	if mode == ModeJunqi {
		return junqiBoard[pos.Row][pos.Col]
	}
	return defaultBoard[pos.Row][pos.Col]
}

func IsPlayable(pos Pos) bool {
	return IsPlayableForMode(ModeSiguo, pos)
}

func IsPlayableForMode(mode GameMode, pos Pos) bool {
	return BoardCellForMode(mode, pos).Type != OffBoard
}

func IsRailroad(pos Pos) bool {
	return IsRailroadForMode(ModeSiguo, pos)
}

func IsRailroadForMode(mode GameMode, pos Pos) bool {
	cell := BoardCellForMode(mode, pos)
	return cell.Type == Railroad || cell.Type == Frontline
}

func buildBoard() [BoardSize][BoardSize]Cell {
	var b [BoardSize][BoardSize]Cell
	for r := range b {
		for c := range b[r] {
			b[r][c] = Cell{Type: OffBoard}
		}
	}

	fillHome := func(seat Seat) {
		for _, p := range homeCells(seat) {
			t := Normal
			if isHomeRail(seat, p) {
				t = Railroad
			}
			b[p.Row][p.Col] = Cell{Type: t, Home: seatPtr(seat)}
		}
		for _, p := range campCells(seat) {
			b[p.Row][p.Col] = Cell{Type: Camp, Home: seatPtr(seat)}
		}
		for _, p := range hqCells(seat) {
			b[p.Row][p.Col] = Cell{Type: HQ, Home: seatPtr(seat)}
		}
	}

	fillHome(North)
	fillHome(East)
	fillHome(South)
	fillHome(West)

	for _, row := range []int{6, 8, 10} {
		for _, col := range []int{6, 8, 10} {
			b[row][col] = Cell{Type: Frontline}
		}
	}

	return b
}

func buildJunqiBoard() [BoardSize][BoardSize]Cell {
	var b [BoardSize][BoardSize]Cell
	for r := range b {
		for c := range b[r] {
			b[r][c] = Cell{Type: OffBoard}
		}
	}

	for r := 2; r <= 7; r++ {
		for c := 6; c <= 10; c++ {
			b[r][c] = Cell{Type: Normal, Home: seatPtr(North)}
		}
	}
	for r := 9; r <= 14; r++ {
		for c := 6; c <= 10; c++ {
			b[r][c] = Cell{Type: Normal, Home: seatPtr(South)}
		}
	}
	for _, p := range junqiRailCells() {
		var home *Seat
		switch {
		case p.Row <= 7:
			home = seatPtr(North)
		case p.Row >= 9:
			home = seatPtr(South)
		}
		t := Railroad
		if p.Row == 8 && p.Col == 8 {
			t = Frontline
		}
		b[p.Row][p.Col] = Cell{Type: t, Home: home}
	}
	for _, col := range []int{6, 8, 10} {
		t := Railroad
		if col == 8 {
			t = Frontline
		}
		b[8][col] = Cell{Type: t}
	}
	for _, p := range junqiCampCells(North) {
		b[p.Row][p.Col] = Cell{Type: Camp, Home: seatPtr(North)}
	}
	for _, p := range junqiCampCells(South) {
		b[p.Row][p.Col] = Cell{Type: Camp, Home: seatPtr(South)}
	}
	for _, p := range junqiHQCells(North) {
		b[p.Row][p.Col] = Cell{Type: HQ, Home: seatPtr(North)}
	}
	for _, p := range junqiHQCells(South) {
		b[p.Row][p.Col] = Cell{Type: HQ, Home: seatPtr(South)}
	}
	return b
}

func seatPtr(s Seat) *Seat {
	v := s
	return &v
}

func homeCells(seat Seat) []Pos {
	var out []Pos
	switch seat {
	case North:
		for r := 0; r <= 5; r++ {
			for c := 6; c <= 10; c++ {
				out = append(out, Pos{r, c})
			}
		}
	case South:
		for r := 11; r <= 16; r++ {
			for c := 6; c <= 10; c++ {
				out = append(out, Pos{r, c})
			}
		}
	case West:
		for r := 6; r <= 10; r++ {
			for c := 0; c <= 5; c++ {
				out = append(out, Pos{r, c})
			}
		}
	case East:
		for r := 6; r <= 10; r++ {
			for c := 11; c <= 16; c++ {
				out = append(out, Pos{r, c})
			}
		}
	}
	return out
}

func campCells(seat Seat) []Pos {
	switch seat {
	case North:
		return []Pos{{2, 7}, {2, 9}, {3, 8}, {4, 7}, {4, 9}}
	case South:
		return []Pos{{12, 7}, {12, 9}, {13, 8}, {14, 7}, {14, 9}}
	case West:
		return []Pos{{7, 2}, {7, 4}, {8, 3}, {9, 2}, {9, 4}}
	case East:
		return []Pos{{7, 12}, {7, 14}, {8, 13}, {9, 12}, {9, 14}}
	default:
		return nil
	}
}

func hqCells(seat Seat) []Pos {
	switch seat {
	case North:
		return []Pos{{0, 7}, {0, 9}}
	case South:
		return []Pos{{16, 7}, {16, 9}}
	case West:
		return []Pos{{7, 0}, {9, 0}}
	case East:
		return []Pos{{7, 16}, {9, 16}}
	default:
		return nil
	}
}

func junqiCampCells(seat Seat) []Pos {
	switch seat {
	case North:
		return []Pos{{4, 7}, {4, 9}, {5, 8}, {6, 7}, {6, 9}}
	case South:
		return []Pos{{10, 7}, {10, 9}, {11, 8}, {12, 7}, {12, 9}}
	default:
		return nil
	}
}

func junqiHQCells(seat Seat) []Pos {
	switch seat {
	case North:
		return []Pos{{2, 7}, {2, 9}}
	case South:
		return []Pos{{14, 7}, {14, 9}}
	default:
		return nil
	}
}

func junqiRailCells() []Pos {
	seen := map[Pos]bool{}
	add := func(p Pos) {
		seen[p] = true
	}
	for _, row := range []int{3, 7, 9, 13} {
		for col := 6; col <= 10; col++ {
			add(Pos{row, col})
		}
	}
	for _, col := range []int{6, 8, 10} {
		for row := 3; row <= 13; row++ {
			add(Pos{row, col})
		}
	}
	out := make([]Pos, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	return out
}

func isHomeRail(seat Seat, p Pos) bool {
	switch seat {
	case North:
		return p.Row == 1 || p.Row == 5 || p.Col == 6 || p.Col == 10
	case South:
		return p.Row == 11 || p.Row == 15 || p.Col == 6 || p.Col == 10
	case West:
		return p.Col == 1 || p.Col == 5 || p.Row == 6 || p.Row == 10
	case East:
		return p.Col == 11 || p.Col == 15 || p.Row == 6 || p.Row == 10
	default:
		return false
	}
}

func buildRailroadAdjacency(b [BoardSize][BoardSize]Cell) map[Pos][]Pos {
	adj := buildOrthogonalRailroadAdjacency(b)
	for _, edge := range centralRailEdges() {
		addRailEdge(b, adj, edge[0], edge[1])
		addRailEdge(b, adj, edge[1], edge[0])
	}
	for _, edge := range directConnectorEdges() {
		addRailEdge(b, adj, edge[0], edge[1])
		addRailEdge(b, adj, edge[1], edge[0])
	}
	return adj
}

func buildOrthogonalRailroadAdjacency(b [BoardSize][BoardSize]Cell) map[Pos][]Pos {
	adj := make(map[Pos][]Pos)
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			p := Pos{r, c}
			if !isRailroadCell(b, p) {
				continue
			}
			for _, d := range orthogonalDirs {
				n := Pos{p.Row + d.Row, p.Col + d.Col}
				if n.InBounds() && isRailroadCell(b, n) {
					addRailEdge(b, adj, p, n)
				}
			}
		}
	}
	return adj
}

func addRailEdge(b [BoardSize][BoardSize]Cell, adj map[Pos][]Pos, a, c Pos) {
	if !isRailroadCell(b, a) || !isRailroadCell(b, c) {
		return
	}
	adj[a] = append(adj[a], c)
}

func isRailroadCell(b [BoardSize][BoardSize]Cell, p Pos) bool {
	t := b[p.Row][p.Col].Type
	return t == Railroad || t == Frontline
}

var orthogonalDirs = []Pos{{-1, 0}, {0, 1}, {1, 0}, {0, -1}}
var diagonalDirs = []Pos{{-1, -1}, {-1, 1}, {1, 1}, {1, -1}}

func railConnectorPairs() map[Pos]Pos {
	pairs := map[Pos]Pos{}
	for _, edge := range directConnectorEdges() {
		pairs[edge[0]] = edge[1]
		pairs[edge[1]] = edge[0]
	}
	return pairs
}

func directConnectorEdges() [][2]Pos {
	return [][2]Pos{
		{{5, 6}, {6, 5}},
		{{5, 10}, {6, 11}},
		{{11, 6}, {10, 5}},
		{{11, 10}, {10, 11}},
	}
}

func centralRailEdges() [][2]Pos {
	var edges [][2]Pos
	for _, col := range []int{6, 8, 10} {
		edges = append(edges, [2]Pos{{5, col}, {6, col}})
		edges = append(edges, [2]Pos{{6, col}, {8, col}})
		edges = append(edges, [2]Pos{{8, col}, {10, col}})
		edges = append(edges, [2]Pos{{10, col}, {11, col}})
	}
	for _, row := range []int{6, 8, 10} {
		edges = append(edges, [2]Pos{{row, 5}, {row, 6}})
		edges = append(edges, [2]Pos{{row, 6}, {row, 8}})
		edges = append(edges, [2]Pos{{row, 8}, {row, 10}})
		edges = append(edges, [2]Pos{{row, 10}, {row, 11}})
	}
	return edges
}
