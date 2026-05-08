package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"siguo/internal/game"
)

func TestSeatSlicesMarshalAsNumericArraysInProtocolMessages(t *testing.T) {
	message := ServerMessage{
		Type: "state",
		Room: &RoomSnapshot{
			Code:   "TEAM22",
			Mode:   game.ModeSiguo,
			Phase:  "ended",
			Turn:   game.North,
			Winner: []game.Seat{game.North, game.South},
			Ready:  []game.Seat{game.East},
			Request: &PendingRequest{
				Kind:  "tie",
				From:  game.North,
				Stage: "rivals",
				Acks:  []game.Seat{game.South},
			},
			View: &game.ClientView{
				Mode:   game.ModeSiguo,
				Viewer: game.North,
				Phase:  game.Ended,
				Turn:   game.North,
				Winner: []game.Seat{game.North, game.South},
			},
		},
		Event: &game.Event{
			Type:    game.EventGameEnded,
			Winners: []game.Seat{game.North, game.South},
			Losers:  []game.Seat{game.East, game.West},
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	jsonText := string(data)
	for _, want := range []string{
		`"winner":[0,2]`,
		`"ready":[1]`,
		`"acks":[2]`,
		`"Winner":[0,2]`,
		`"Winners":[0,2]`,
		`"Losers":[1,3]`,
	} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("protocol JSON = %s, missing %s", jsonText, want)
		}
	}
	if strings.Contains(jsonText, `"AAI="`) || strings.Contains(jsonText, `"AQM="`) {
		t.Fatalf("protocol JSON contains base64-encoded seat slices: %s", jsonText)
	}
}
