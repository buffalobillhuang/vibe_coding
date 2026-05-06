package game

type CombatResult struct {
	Outcome       CombatOutcome
	AttackerDies  bool
	DefenderDies  bool
	FlagCaptured  bool
	CommanderDied []Seat
}

func ResolveCombat(attacker, defender Piece) CombatResult {
	result := CombatResult{}

	switch {
	case attacker.Rank == Bomb || defender.Rank == Bomb:
		result.Outcome = BothDie
		result.AttackerDies = true
		result.DefenderDies = true
	case defender.Rank == Flag:
		result.Outcome = AttackerWins
		result.DefenderDies = true
		result.FlagCaptured = true
	case attacker.Rank == Engineer && defender.Rank == Mine:
		result.Outcome = AttackerWins
		result.DefenderDies = true
	case defender.Rank == Mine:
		result.Outcome = DefenderWins
		result.AttackerDies = true
	case attacker.Rank.Strength() > defender.Rank.Strength():
		result.Outcome = AttackerWins
		result.DefenderDies = true
	case attacker.Rank.Strength() < defender.Rank.Strength():
		result.Outcome = DefenderWins
		result.AttackerDies = true
	default:
		result.Outcome = BothDie
		result.AttackerDies = true
		result.DefenderDies = true
	}

	if result.AttackerDies && attacker.Rank == Commander {
		result.CommanderDied = append(result.CommanderDied, attacker.Owner)
	}
	if result.DefenderDies && defender.Rank == Commander {
		result.CommanderDied = append(result.CommanderDied, defender.Owner)
	}

	return result
}
