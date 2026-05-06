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
var railroadAdj map[Pos][]Pos

func init() {
	defaultBoard = buildBoard()
	railroadAdj = buildRailroadAdjacency(defaultBoard)
}

func BoardCell(pos Pos) Cell {
	if !pos.InBounds() {
		return Cell{Type: OffBoard}
	}
	return defaultBoard[pos.Row][pos.Col]
}

func IsPlayable(pos Pos) bool {
	return BoardCell(pos).Type != OffBoard
}

func IsRailroad(pos Pos) bool {
	cell := BoardCell(pos)
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
	adj := make(map[Pos][]Pos)
	add := func(a, c Pos) {
		if !isRailroadCell(b, a) || !isRailroadCell(b, c) {
			return
		}
		adj[a] = append(adj[a], c)
	}
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			p := Pos{r, c}
			if !isRailroadCell(b, p) {
				continue
			}
			for _, d := range orthogonalDirs {
				n := Pos{p.Row + d.Row, p.Col + d.Col}
				if n.InBounds() && isRailroadCell(b, n) {
					add(p, n)
				}
			}
		}
	}
	for _, edge := range centralRailEdges() {
		add(edge[0], edge[1])
		add(edge[1], edge[0])
	}
	for _, edge := range directConnectorEdges() {
		add(edge[0], edge[1])
		add(edge[1], edge[0])
	}
	return adj
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
