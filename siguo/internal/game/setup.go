package game

import "fmt"

var StandardArmy = map[Rank]int{
	Flag:               1,
	Mine:               3,
	Bomb:               2,
	Engineer:           3,
	PlatoonLeader:      3,
	CompanyCommander:   3,
	BattalionCommander: 2,
	RegimentCommander:  2,
	BrigadeCommander:   2,
	DivisionCommander:  2,
	CorpsCommander:     1,
	Commander:          1,
}

func ValidateSetup(owner Seat, pieces []Piece, placements map[PieceID]Pos) error {
	counts := map[Rank]int{}
	occupied := map[Pos]bool{}

	for _, piece := range pieces {
		if piece.Owner != owner {
			continue
		}
		pos, ok := placements[piece.ID]
		if !ok {
			return fmt.Errorf("piece %d has no placement", piece.ID)
		}
		if occupied[pos] {
			return fmt.Errorf("multiple pieces placed at %v", pos)
		}
		occupied[pos] = true

		if err := validateSetupCell(owner, piece.Rank, pos); err != nil {
			return fmt.Errorf("piece %d: %w", piece.ID, err)
		}
		counts[piece.Rank]++
	}

	for rank, want := range StandardArmy {
		if counts[rank] != want {
			return fmt.Errorf("%s count = %d, want %d", rank, counts[rank], want)
		}
	}
	for rank, got := range counts {
		if _, ok := StandardArmy[rank]; !ok {
			return fmt.Errorf("%s is not part of the standard army", rank)
		}
		if got > StandardArmy[rank] {
			return fmt.Errorf("%s count = %d, want %d", rank, got, StandardArmy[rank])
		}
	}

	return nil
}

func validateSetupCell(owner Seat, rank Rank, pos Pos) error {
	return ValidateSetupCell(owner, rank, pos)
}

func ValidateSetupCell(owner Seat, rank Rank, pos Pos) error {
	cell := BoardCell(pos)
	if cell.Type == OffBoard {
		return ErrDestination
	}
	if cell.Home == nil || *cell.Home != owner {
		return fmt.Errorf("position %v is outside %s home zone", pos, owner)
	}
	if cell.Type == Camp {
		return fmt.Errorf("position %v is a camp", pos)
	}
	if rank == Flag && cell.Type != HQ {
		return fmt.Errorf("flag must be placed in HQ")
	}
	if rank == Mine && !isBackTwoRows(owner, pos) {
		return fmt.Errorf("mine must be placed in the back two rows")
	}
	if rank == Bomb && isFrontRow(owner, pos) {
		return fmt.Errorf("bomb cannot be placed on the front row")
	}
	return nil
}

func isBackTwoRows(owner Seat, pos Pos) bool {
	switch owner {
	case North:
		return pos.Row <= 1
	case South:
		return pos.Row >= 15
	case West:
		return pos.Col <= 1
	case East:
		return pos.Col >= 15
	default:
		return false
	}
}

func isFrontRow(owner Seat, pos Pos) bool {
	switch owner {
	case North:
		return pos.Row == 5
	case South:
		return pos.Row == 11
	case West:
		return pos.Col == 5
	case East:
		return pos.Col == 11
	default:
		return false
	}
}
