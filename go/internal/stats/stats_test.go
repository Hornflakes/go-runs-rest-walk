package stats

import (
	"testing"
)

func TestGameFrameStatsString(t *testing.T) {
	gfs := NewGameFrameStats()
	gfs.AddDeltaTime(16_100)
	gfs.AddDeltaTime(18_000)

	want := "1,1,0,0,0,0,0,0"
	if got := gfs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGameFrameStatsAddDeltaTime(t *testing.T) {
	gfs := NewGameFrameStats()
	gfs.AddDeltaTime(16_000)
	gfs.AddDeltaTime(18_000)
	gfs.AddDeltaTime(20_000)
	gfs.AddDeltaTime(22_000)
	gfs.AddDeltaTime(24_000)
	gfs.AddDeltaTime(26_000)
	gfs.AddDeltaTime(31_000)
	gfs.AddDeltaTime(41_000)

	want := "1,1,1,1,1,1,1,1"
	if got := gfs.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
