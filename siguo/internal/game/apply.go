package game

func ApplyMove(state *GameState, move Move) (*GameState, []Event, error) {
	if state.Phase != Playing {
		return nil, nil, ErrGameNotPlaying
	}
	piece, ok := state.Pieces[move.PieceID]
	if !ok {
		return nil, nil, ErrInvalidPiece
	}
	if !piece.Alive {
		return nil, nil, ErrPieceNotAlive
	}
	if !piece.Rank.Movable() {
		return nil, nil, ErrPieceImmobile
	}
	if state.Turn != piece.Owner {
		return nil, nil, ErrNotPieceTurn
	}
	from, ok := state.positionOf(move.PieceID)
	if !ok || from != move.From {
		return nil, nil, ErrInvalidMove
	}
	legal, ok := legalMoveFor(state, move)
	if !ok {
		return nil, nil, ErrInvalidMove
	}
	eventPath := eventPathForMove(legal)

	next := state.Clone()
	var events []Event

	defender, occupied := next.PieceAt(move.To)
	if !occupied {
		next.movePiece(move.PieceID, move.To)
		events = append(events, Event{Type: EventMove, From: move.From, To: move.To, Path: eventPath, Attacker: move.PieceID})
	} else {
		result := ResolveCombat(piece, defender)

		if result.AttackerDies {
			next.kill(piece.ID)
		} else {
			next.movePiece(piece.ID, move.To)
		}
		if result.DefenderDies {
			next.kill(defender.ID)
		}
		if result.Outcome == AttackerWins && !result.AttackerDies {
			next.BoardIdx[move.To.Row][move.To.Col] = piece.ID
			next.Positions[piece.ID] = move.To
		}

		events = append(events, Event{
			Type:     EventCombat,
			From:     move.From,
			To:       move.To,
			Path:     eventPath,
			Attacker: piece.ID,
			Defender: defender.ID,
			Outcome:  result.Outcome,
		})

		for _, owner := range result.CommanderDied {
			if revealed, _ := next.revealFlag(owner); revealed {
				events = append(events, Event{Type: EventFlagRevealed, Owner: owner})
			}
		}

		if result.FlagCaptured {
			next.eliminate(defender.Owner)
			events = append(events, Event{Type: EventFlagCaptured, Owner: defender.Owner})
		}
	}

	if moved, ok := next.Pieces[piece.ID]; ok && moved.Alive && moved.Rank == Engineer && next.boardCell(move.To).Type == Mountain {
		next.Revealed[piece.ID] = true
	}

	if winners := next.checkWinner(); winners != nil {
		next.Phase = Ended
		next.Winner = winners
		events = append(events, Event{Type: EventTeamEliminated, Losers: losingSide(next.Mode, winners), Winners: winners})
		events = append(events, Event{Type: EventGameEnded, Winners: winners})
	} else if winners := next.noMoveWinner(); winners != nil {
		next.Phase = Ended
		next.Winner = winners
		events = append(events, Event{Type: EventGameEnded, Winners: winners})
	} else {
		next.advanceTurn()
	}

	next.History = append(next.History, events...)
	return next, events, nil
}

func legalMoveFor(state *GameState, move Move) (Move, bool) {
	for _, legal := range LegalMoves(state, move.PieceID) {
		if legal.To == move.To {
			return legal, true
		}
	}
	return Move{}, false
}

func eventPathForMove(move Move) []Pos {
	if len(move.Path) > 0 {
		return append([]Pos(nil), move.Path...)
	}
	return []Pos{move.From, move.To}
}

func losingTeam(winners []Seat) []Seat {
	return losingSide(ModeSiguo, winners)
}

func losingSide(mode GameMode, winners []Seat) []Seat {
	if mode == ModeJunqi {
		if len(winners) != 1 {
			return nil
		}
		if winners[0] == North {
			return []Seat{South}
		}
		return []Seat{North}
	}
	if len(winners) != 2 {
		return nil
	}
	if winners[0].SameTeam(North) {
		return []Seat{East, West}
	}
	return []Seat{North, South}
}
