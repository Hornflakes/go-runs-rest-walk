package server

import "testing"

// Clients send {"type": 2} to shoot. If this changes, the game loop
// will stop reacting to shooting.
func TestUnmarshalShoot(t *testing.T) {
	m, err := UnmarshalMessage([]byte(`{"type": 2}`))
	if err != nil {
		t.Fatal(err)
	}
	if m.Type != Shoot {
		t.Fatalf("got %d, want %d", m.Type, Shoot)
	}
}
