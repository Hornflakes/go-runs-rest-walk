package gameloop

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type SyntheticClock struct {
	now time.Time
}

func (sc *SyntheticClock) Now() time.Time {
	return sc.now
}

func (sc *SyntheticClock) SetNow(now time.Time) {
	sc.now = now
}
