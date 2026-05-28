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
	if m.Type != Shoot {
		t.Errorf("Type = %d, want %d", m.Type, Shoot)
	}
}

func TestCreateWinnerMessage(t *testing.T) {
	gfs := stats.NewGameFrameStats()
	gfs.AddDeltaTime(16_000)
	gfs.AddDeltaTime(18_000)

	msg := CreateWinnerMessage(gfs).Message

	if msg.Type != GameOver {
		t.Errorf("Type = %d, want %d", msg.Type, GameOver)
	}

	want := "winner(0)->1,1,0,0,0,0,0,0"
	if got := msg.Msg; got != want {
		t.Errorf("Msg = %q, want %q", got, want)
	}
}
