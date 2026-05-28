package stats

import (
	"fmt"
	"strings"
	"sync"
)

var ActiveGames uint
var mutex sync.Mutex

func AddActiveGame() {
	mutex.Lock()
	ActiveGames += 1
	mutex.Unlock()
}

func RemoveActiveGame() {
	mutex.Lock()
	ActiveGames -= 1
	mutex.Unlock()
}

type GameFrameStats struct {
	FrameBuckets [8]int64
}

func NewGameFrameStats() *GameFrameStats {
	return &GameFrameStats{
		FrameBuckets: [8]int64{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (gfs *GameFrameStats) String() string {
	out := make([]string, 8)
	for idx, num := range gfs.FrameBuckets {
		out[idx] = fmt.Sprintf("%d", num)
	}
	return strings.Join(out, ",")
}

func (gfs *GameFrameStats) AddDeltaTime(delta int64) {
	if delta > 40_999 {
		gfs.FrameBuckets[7] += 1
	} else if delta > 30_999 {
		gfs.FrameBuckets[6] += 1
	} else if delta > 25_999 {
		gfs.FrameBuckets[5] += 1
	} else if delta > 23_999 {
		gfs.FrameBuckets[4] += 1
	} else if delta > 21_999 {
		gfs.FrameBuckets[3] += 1
	} else if delta > 19_999 {
		gfs.FrameBuckets[2] += 1
	} else if delta > 17_999 {
		gfs.FrameBuckets[1] += 1
	} else {
		gfs.FrameBuckets[0] += 1
	}
}
