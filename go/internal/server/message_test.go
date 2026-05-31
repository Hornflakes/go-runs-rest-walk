package server

import (
	"testing"

	"github.com/hornflakes/go-runs-rest-walk/internal/stats"
)

// Clients send {"type": 3} to shoot. If this changes, the game loop
// will stop reacting to shooting.
func TestUnmarshalShoot(t *testing.T) {
	m, err := UnmarshalMessage([]byte(`{"type": 3}`))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := m.Type, Shoot; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}
}

func TestCreateMessageTypes(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{"Hello", CreateHelloMessage(3).Message.Type, Hello},
		{"Ready", CreateReadyMessage(7).Message.Type, Ready},
		{"GameOn", CreateSocketMessage(GameOn).Message.Type, GameOn},
		{"GameOver winner", CreateWinnerMessage(stats.NewGameFrameStats()).Message.Type, GameOver},
		{"GameOver loser", CreateLoserMessage().Message.Type, GameOver},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Type = %d, want %d", tt.got, tt.want)
			}
		})
	}
}

func TestCreateHelloMessage(t *testing.T) {
	msg := CreateHelloMessage(3).Message

	if got, want := msg.Type, Hello; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}

	if got, want := msg.Msg, "playerId=3"; got != want {
		t.Errorf("Msg = %q, want %q", got, want)
	}
}

func TestCreateReadyMessage(t *testing.T) {
	msg := CreateReadyMessage(7).Message

	if got, want := msg.Type, Ready; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}

	if got, want := msg.Msg, "enemyId=7"; got != want {
		t.Errorf("Msg = %q, want %q", got, want)
	}
}

func TestCreateWinnerMessage(t *testing.T) {
	gfs := stats.NewGameFrameStats()
	gfs.AddDeltaTime(16_000)
	gfs.AddDeltaTime(18_000)

	msg := CreateWinnerMessage(gfs).Message

	if got, want := msg.Type, GameOver; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}

	if got, want := msg.Msg, "winner(0)->1,1,0,0,0,0,0,0"; got != want {
		t.Errorf("Msg = %q, want %q", got, want)
	}
}

func TestCreateLoserMessage(t *testing.T) {
	msg := CreateLoserMessage().Message

	if got, want := msg.Type, GameOver; got != want {
		t.Errorf("Type = %d, want %d", got, want)
	}

	if got, want := msg.Msg, "loser"; got != want {
		t.Errorf("Msg = %q, want %q", got, want)
	}
}

func TestParseHelloMessage(t *testing.T) {
	msg := "playerId=3"

	got, err := ParseHelloMessage(msg)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := got, uint64(3); got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	msg = "3"

	got, err = ParseHelloMessage(msg)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := got, uint64(3); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestParseReadyMessage(t *testing.T) {
	msg := "enemyId=7"

	got, err := ParseReadyMessage(msg)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := got, uint64(7); got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	msg = "7"

	got, err = ParseReadyMessage(msg)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := got, uint64(7); got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
