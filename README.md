# go-runs-rest-walk

Concurrent WebSocket game server benchmark (Go vs TypeScript vs Java).

## Methodology

1. **Environment:** WSL/Linux. Pin **server and load** to the **same CPU** (`taskset -c 0`).
2. **Warmup:** do one short load before benchmarking (required for **Java** JIT; optional but good for Go/TS - thermals, OS cache).
3. **Measure** CPU/RAM usage with `scripts/measure.sh` during the run.
4. **Load:** 10.000 connections x 20 games = 100.000 games; with 50ms stagger between connections; and 40 shards (250 conns/shard) 5ms interval burst, `-connections 10000 -games 20 -stagger 50 -burst`.
5. **Compare** rolling bucket0 at saturation (~2500–3500 active games), not overall ratio alone.
6. **Optional:** Repeat thrice (`run1`, `run2`, `run3`) on the same server process.

Valid run: `load finished | games_over=100000 games_failed=~0` and all clients done.

## Commands (and their order)

```bash
mkdir -p results
```

### 1. Server

```bash
cd go && taskset -c 0 go run ./cmd/server/

cd ts && taskset -c 0 npm run start

cd java && taskset -c 0 ./gradlew run --console=plain
```

#### 1.1 Warmup

```bash
cd go && taskset -c 0 go run ./cmd/load/ \
  -connections 2000 -games 5 -stagger 50 -burst
```

### 2. Measure CPU/RAM (parallel with Load)

```bash
PID=$(lsof -t -i :37373 -sTCP:LISTEN)

./scripts/measure.sh "$PID" results/go-N10000-G20-run1-measure.csv
```

### 3. Load

Change language/tag and `run1` -> `run2` / `run3`:

```bash
cd go && taskset -c 0 go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/go-N10000-G20-run1-load.res
```

Examples: `go-N10000-G20-run1`, `ts-N10000-G20-run1`, `java-N10000-G20-run1`.

### 4. Parse and plot

```bash
cd scripts && go run ./load-parse-results/main.go \
  ../results/go-N10000-G20-run1-load.res

cd scripts && python3 benchmark.py ../results --x active_games
cd scripts && python3 benchmark.py ../results --x window_end
```

Parsing writes `*-rolling.csv` and appends a row to `runs.csv`.

## Results layout

```text
results/
  x-load.res               # load raw stdout  -> winner lines + load heartbeats/finished
  x-rolling.csv            # load-parse-results -> rolling active_games, bucket0_ratio mean
  x-measure.csv            # measure -> pidstat CPU/MEM samples
  runs.csv                 # load-parse-results -> one summary row per parsed run
  rolling-active-games.png # benchmark.py ../results --x active_games
  rolling-over-time.png    # benchmark.py ../results --x window_end
```

x = `{lang}-N{connections}-G{games}[-tag]-run{n}`
