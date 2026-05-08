package game

import (
	"encoding/json"
	"testing"
)

func TestSeatSlicesMarshalAsJSONNumbers(t *testing.T) {
	data, err := json.Marshal(struct {
		Winner []Seat `json:"winner"`
	}{Winner: []Seat{North, South}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(data) != `{"winner":[0,2]}` {
		t.Fatalf("json = %s, want numeric winner seats", data)
	}
}
