package game

import "fmt"

const BoardSize = 17

type GameMode string

const (
	ModeSiguo GameMode = "siguo"
	ModeJunqi GameMode = "junqi"
)

func (m GameMode) String() string {
	if m == ModeJunqi {
		return string(ModeJunqi)
	}
	return string(ModeSiguo)
}

type Seat uint8

const (
	North Seat = iota
	East
	South
	West
)

var Seats = []Seat{North, East, South, West}

func ActiveSeats(mode GameMode) []Seat {
	if mode == ModeJunqi {
		return []Seat{North, South}
	}
	return []Seat{North, East, South, West}
}

func TurnOrder(mode GameMode) []Seat {
	if mode == ModeJunqi {
		return []Seat{North, South}
	}
	return []Seat{North, West, South, East}
}

func (s Seat) String() string {
	switch s {
	case North:
		return "North"
	case East:
		return "East"
	case South:
		return "South"
	case West:
		return "West"
	default:
		return fmt.Sprintf("Seat(%d)", s)
	}
}

func (s Seat) Partner() Seat {
	switch s {
	case North:
		return South
	case South:
		return North
	case East:
		return West
	case West:
		return East
	default:
		return s
	}
}

func (s Seat) SameTeam(other Seat) bool {
	return s == other || s.Partner() == other
}

func (s Seat) Next() Seat {
	return Seat((uint8(s) + 3) % 4)
}

type Rank uint8

const (
	Unknown Rank = iota
	Flag
	Mine
	Bomb
	Engineer
	PlatoonLeader
	CompanyCommander
	BattalionCommander
	RegimentCommander
	BrigadeCommander
	DivisionCommander
	CorpsCommander
	Commander
)

func (r Rank) String() string {
	switch r {
	case Unknown:
		return "Unknown"
	case Flag:
		return "军旗"
	case Mine:
		return "地雷"
	case Bomb:
		return "炸弹"
	case Engineer:
		return "工兵"
	case PlatoonLeader:
		return "排长"
	case CompanyCommander:
		return "连长"
	case BattalionCommander:
		return "营长"
	case RegimentCommander:
		return "团长"
	case BrigadeCommander:
		return "旅长"
	case DivisionCommander:
		return "师长"
	case CorpsCommander:
		return "军长"
	case Commander:
		return "司令"
	default:
		return fmt.Sprintf("Rank(%d)", r)
	}
}

func (r Rank) Movable() bool {
	return r != Flag && r != Mine && r != Unknown
}

func (r Rank) Strength() int {
	switch r {
	case Engineer:
		return 1
	case PlatoonLeader:
		return 2
	case CompanyCommander:
		return 3
	case BattalionCommander:
		return 4
	case RegimentCommander:
		return 5
	case BrigadeCommander:
		return 6
	case DivisionCommander:
		return 7
	case CorpsCommander:
		return 8
	case Commander:
		return 9
	default:
		return 0
	}
}

type PieceID uint32

type Piece struct {
	ID    PieceID
	Owner Seat
	Rank  Rank
	Alive bool
}

type Pos struct {
	Row int
	Col int
}

func (p Pos) InBounds() bool {
	return p.Row >= 0 && p.Row < BoardSize && p.Col >= 0 && p.Col < BoardSize
}

type Phase uint8

const (
	Setup Phase = iota
	Playing
	Ended
)

type Clock struct {
	MSRemaining int64
}

type Move struct {
	PieceID PieceID
	From    Pos
	To      Pos
	Path    []Pos
}

type EventType string

const (
	EventMove           EventType = "move"
	EventCombat         EventType = "combat"
	EventFlagRevealed   EventType = "flagRevealed"
	EventFlagCaptured   EventType = "flagCaptured"
	EventTeamEliminated EventType = "teamEliminated"
	EventGameEnded      EventType = "gameEnded"
)

type CombatOutcome string

const (
	AttackerWins CombatOutcome = "attackerWins"
	DefenderWins CombatOutcome = "defenderWins"
	BothDie      CombatOutcome = "bothDie"
)

type Event struct {
	Type     EventType
	From     Pos
	To       Pos
	Attacker PieceID
	Defender PieceID
	Outcome  CombatOutcome
	Owner    Seat
	Losers   []Seat
	Winners  []Seat
}

type GameState struct {
	Mode       GameMode
	Phase      Phase
	Turn       Seat
	Pieces     map[PieceID]Piece
	Positions  map[PieceID]Pos
	BoardIdx   [BoardSize][BoardSize]PieceID
	Clocks     map[Seat]Clock
	Revealed   map[PieceID]bool
	Eliminated map[Seat]bool
	Winner     []Seat
	History    []Event
}
