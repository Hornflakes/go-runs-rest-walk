package server

import (
	"testing"

	"github.com/hornflakes/go-runs-rest-walk/internal/stats"
)

// Clients send {"type": 2} to shoot. If this changes, the game loop
// will stop reacting to shooting.
func TestUnmarshalShoot(t *testing.T) {
	m, err := UnmarshalMessage([]byte(`{"type": 2}`))
	if err != nil {
		t.Fatal(err)
	}

	want := Shoot
	if got := m.Type; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}
}

func TestCreateWinnerMessage(t *testing.T) {
	gfs := stats.NewGameFrameStats()
	gfs.AddDeltaTime(16_000)
	gfs.AddDeltaTime(18_000)

	msg := CreateWinnerMessage(gfs).Message

	want0 := GameOver
	if got := msg.Type; got != want0 {
		t.Errorf("Type = %d, want %d", got, want0)
	}

	want1 := "winner(0)->1,1,0,0,0,0,0,0"
	if got := msg.Msg; got != want1 {
		t.Errorf("Msg = %q, want %q", got, want1)
	}
}

func TestCreateLoserMessage(t *testing.T) {
	msg := CreateLoserMessage().Message

	want0 := GameOver
	if got := msg.Type; got != want0 {
		t.Errorf("Type = %d, want %d", got, want0)
	}

	want1 := "loser"
	if got := msg.Msg; got != want1 {
		t.Errorf("Msg = %q, want %q", got, want1)
	}
}
