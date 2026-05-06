package game

import "testing"

func TestResolveCombatTruthTableHighlights(t *testing.T) {
	tests := []struct {
		name          string
		attacker      Rank
		defender      Rank
		outcome       CombatOutcome
		attackerDies  bool
		defenderDies  bool
		flagCaptured  bool
		commanderDied int
	}{
		{
			name:          "bomb kills both",
			attacker:      Bomb,
			defender:      Commander,
			outcome:       BothDie,
			attackerDies:  true,
			defenderDies:  true,
			commanderDied: 1,
		},
		{
			name:         "engineer defuses mine",
			attacker:     Engineer,
			defender:     Mine,
			outcome:      AttackerWins,
			defenderDies: true,
		},
		{
			name:          "mine kills non engineer",
			attacker:      Commander,
			defender:      Mine,
			outcome:       DefenderWins,
			attackerDies:  true,
			commanderDied: 1,
		},
		{
			name:         "flag is captured",
			attacker:     Engineer,
			defender:     Flag,
			outcome:      AttackerWins,
			defenderDies: true,
			flagCaptured: true,
		},
		{
			name:          "commander death tracked",
			attacker:      Commander,
			defender:      Bomb,
			outcome:       BothDie,
			attackerDies:  true,
			defenderDies:  true,
			commanderDied: 1,
		},
		{
			name:         "higher rank wins",
			attacker:     CorpsCommander,
			defender:     DivisionCommander,
			outcome:      AttackerWins,
			defenderDies: true,
		},
		{
			name:         "equal rank both die",
			attacker:     RegimentCommander,
			defender:     RegimentCommander,
			outcome:      BothDie,
			attackerDies: true,
			defenderDies: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveCombat(
				Piece{ID: 1, Owner: North, Rank: tt.attacker, Alive: true},
				Piece{ID: 2, Owner: East, Rank: tt.defender, Alive: true},
			)
			if got.Outcome != tt.outcome {
				t.Fatalf("outcome = %s, want %s", got.Outcome, tt.outcome)
			}
			if got.AttackerDies != tt.attackerDies {
				t.Fatalf("attacker dies = %v, want %v", got.AttackerDies, tt.attackerDies)
			}
			if got.DefenderDies != tt.defenderDies {
				t.Fatalf("defender dies = %v, want %v", got.DefenderDies, tt.defenderDies)
			}
			if got.FlagCaptured != tt.flagCaptured {
				t.Fatalf("flag captured = %v, want %v", got.FlagCaptured, tt.flagCaptured)
			}
			if len(got.CommanderDied) != tt.commanderDied {
				t.Fatalf("commander deaths = %d, want %d", len(got.CommanderDied), tt.commanderDied)
			}
		})
	}
}
