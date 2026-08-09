package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
	"github.com/hornflakes/unpocologo"
)

var (
	gamesOver      atomic.Uint64
	gamesFailed    atomic.Uint64
	clientsStarted atomic.Uint64
	clientsDone    atomic.Uint64
)

func main() {
	connections := flag.Int("connections", 10, "number of websockets")
	games := flag.Int("games", 0, "number of games per connection, 0 = run until Ctrl+C")
	host := flag.String("host", "127.0.0.1", "host to connect to")
	port := flag.Int("port", 37373, "port to connect to")
	path := flag.String("path", "/", "path to connect to")
	stagger := flag.Int("stagger", 50, "stagger between connections in ms")
	fire := flag.Int("fire", 200, "fire rate in ms")
	burst := flag.Bool("burst", true, "burst fire loop; ignores -fire")
	burstInterval := flag.Int("burst-interval", defaultBurstIntervalMs, "burst interval in ms")
	burstShards := flag.Int("burst-shards", defaultBurstShardsCount, "burst shard count")
	flag.Parse()

	if *connections < 1 {
		log.Fatal("connections must be at least 1")
	}
	if *path == "" || (*path)[0] != '/' {
		log.Fatal("path must start with /")
	}

	shootMsg, err := json.Marshal(server.CreateMessage(server.Shoot))
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("ws://%s:%d%s", *host, *port, *path)
	staggerIntervalTime := time.Duration(*stagger) * time.Millisecond
	fireIntervalTime := time.Duration(*fire) * time.Millisecond

	fireLabel := fireIntervalTime.String()
	var pool *burstPool
	if *burst {
		pool = newBurstPool(*burstShards, *burstInterval, shootMsg)
		fireLabel = fmt.Sprintf("burst-%dms-%dshards", *burstInterval, *burstShards)
	}

	unpocologo.Milestonef("load", "url=%s connections=%d games=%d stagger=%s fire=%s",
		url, *connections, *games, staggerIntervalTime, fireLabel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if pool != nil {
		go pool.fireLoop(ctx)
	}

	var wg sync.WaitGroup
	for id := 0; id < *connections; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			select {
			case <-time.After(time.Duration(id) * staggerIntervalTime):
			case <-ctx.Done():
				return
			}
			runClient(ctx, url, id, pool, fireIntervalTime, shootMsg, *games)
		}(id)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				unpocologo.Infof("load heartbeat",
					"games_over=%d games_failed=%d clients_started=%d/%d clients_done=%d/%d",
					gamesOver.Load(), gamesFailed.Load(),
					clientsStarted.Load(), *connections,
					clientsDone.Load(), *connections)
			}
		}
	}()

	wg.Wait()

	unpocologo.Milestonef("load finished",
		"games_over=%d games_failed=%d clients_started=%d/%d clients_done=%d/%d",
		gamesOver.Load(), gamesFailed.Load(),
		clientsStarted.Load(), *connections,
		clientsDone.Load(), *connections)
}
