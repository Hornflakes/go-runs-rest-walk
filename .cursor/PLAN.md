# go-runs-rest-walk — project plan

## Repos

| Role                      | Path                                                                       |
| ------------------------- | -------------------------------------------------------------------------- |
| **Implement**             | `go-runs-rest-walk` — `github.com/hornflakes/go-runs-rest-walk`, Go 1.26.3 |
| **Reference (read only)** | `../tyrone-biggums`                                                        |
| **Default port**          | 37373, path `/`                                                            |

Attribution: inspired by tyrone-biggums; own implementation; cite in thesis.

---

## Done

- Go game server (pair → ready → 60 Hz loop → GameOver + histogram)
- `cmd/bot` — single-client dev / smoke test
- `cmd/load` — N connections, stagger, fire, heartbeats, `clients_started` / `clients_done` / `games_failed`
- `cmd/load` — prints winner lines on stdout (`fmt.Println`); `gamesOver` counts winners only
- Disconnect + context (ready phase), `GameQueue.Stop()`, safe sleep
- **`scripts/load-parse-results/`** — parse `.res` → `{run_id}-rolling.csv` + append `results/runs.csv`
- **`scripts/benchmark.py`** — plot `*-rolling.csv` → PNG (`--x window_end` default, `--x active_games` for tyrone load curve)
- End-to-end smoke: capture → parse → plot verified (e.g. `go-N100-G20-run1`)

---

## Wire protocol (actual)

JSON: `{"type": n, "msg"?: "..."}`

| type | Name     | Notes                                                               |
| ---- | -------- | ------------------------------------------------------------------- |
| 0    | Hello    | Server → client                                                     |
| 1    | Ready    | Handshake echo                                                      |
| 2    | Shoot    | Client → server                                                     |
| 3    | GameOver | Winner: `winner=<playerId> histogram=b0,b1,...,b7 active_games=<N>` |
| 3    | GameOver | Loser: `"loser"` (not counted in benchmark)                         |

- `winner=<id>` — winning **player id**
- `active_games` — global concurrent games at game end (not game number)
- Tyrone reference used `winner(activeGames)___b0,...` (not implemented in parser; add if needed for reference files)

Constants: see `internal/gameloop/spec.go` (tick 16 ms, fire rates, etc.).

---

## Benchmark philosophy (tyrone shooter model)

- **Load client** is the benchmark harness (like tyrone `game_player`).
- **Histogram data** comes from **load stdout** (GameOver winner lines), teed to `.res` files via `2>&1 | tee`.
- **Server:** run in a terminal and watch live — **no server log capture**.
- **Same `cmd/load`** against Go and (later) TypeScript servers for fair comparison.
- Per-game detail stays in `.res` (like tyrone `.res` / `.results`); no separate `games.csv`.

**Winner line count:**

```text
winner lines ≈ connections × games_per_conn ÷ 2
```

(two clients per game; only the winner prints a line)

---

## Results layout

Flat `results/` folder. Run-id-first names so `ls` groups artifacts:

```text
results/
  go-N100-G20-run1-load.res
  go-N100-G20-run1-rolling.csv
  runs.csv
  rolling.png              # default: --x window_end
  rolling-load.png         # --x active_games
```

Naming: `{lang}-N{connections}-G{games_per_conn}-run{n}-load.res` → rolling replaces `-load.res` with `-rolling.csv`.

---

## Benchmark workflow

### 1. Run capture

**Formal / thesis runs** — fixed `-connections N` and `-games M`; exit when load prints `load finished`:

```bash
# Terminal 1 — server (watch in terminal, no tee)
cd go && go run ./cmd/server/main.go

# Terminal 2 — load (2>&1 merges logger + winner lines for tee)
cd go && go run ./cmd/load/main.go -connections 100 -games 20 -stagger 25 -fire 200 2>&1 \
  | tee ../results/go-N100-G20-run1-load.res
```

**Dev** — `-games 0` and Ctrl+C when done watching.

Repeat **≥3 runs** per N+G (`run1`, `run2`, `run3`). N ladder e.g. 500 → 1k → 2k → 5k → 8k → 10k (stop when `games_failed` rises or `bucket0_ratio` collapses).

Per run note in thesis/lab book: OS, CPU/RAM, git commit, stagger, any anomalies.

### 2. Parse

```bash
cd scripts && go run ./load-parse-results/main.go ../results/go-N100-G20-run1-load.res
```

**`scripts/load-parse-results/`** (Go, `package main`; module in `scripts/go.mod`)

- Parse winner lines: `winner=(\d+) histogram=([\d,]+) active_games=(\d+)`
- Parse `load | url=...` and `load finished | ...` for `runs.csv`
- Strip ANSI from teed logger lines; ignore heartbeats for CSV output
- Rolling requires `≥ window` winner lines (default window 100); `runs.csv` appended even if rolling fails

**Output: `{run_id}-rolling.csv`** (every 100 games)

```text
window_end,active_games,bucket0_ratio
```

- `bucket0_ratio = sum(w0) / sum(w0..w7)` — **fraction 0–1**, formatted `%.2f`
- `active_games` = mean of `active_games` from winner lines in the window

**Output: append `results/runs.csv`** (one row per parse)

```text
run_id,language,connections,games_per_conn,run_number,url,stagger,fire,
games_parsed,bucket0_ratio_overall,games_over,games_failed,clients_started,clients_done
```

### 3. Plot

```bash
cd scripts && py -3 benchmark.py ../results
py -3 benchmark.py ../results --x active_games   # → rolling-load.png
```

**`scripts/benchmark.py`** (Python, matplotlib)

| Output             | `--x`                  | Question answered                      |
| ------------------ | ---------------------- | -------------------------------------- |
| `rolling.png`      | `window_end` (default) | Tick quality **over run progress**     |
| `rolling-load.png` | `active_games`         | Tyrone-style quality **vs load level** |

Overlay: drop multiple `*-rolling.csv` in `results/` (later `go-...` + `ts-...`).

**Optional summary plot (not implemented):** read `runs.csv` — X = `connections`, Y = median `bucket0_ratio_overall` across run1–3 at each N. One point per N for thesis saturation overview.

### 4. CPU/RAM (optional, tyrone `measure`)

During steady state, sample server + load PIDs (`pidstat` or copy tyrone `measure`).

Separate from histogram pipeline; supports “loader vs server CPU” narrative.

### 5. Profiling (optional)

- `perf` + FlameGraph on server PID at low / mid / saturated N (not pprof)
- Linux/WSL for published numbers

---

## Go vs TypeScript (later)

1. Port TS server to same spec + wire format.
2. Same load flags, same N+G ladder, same scripts.
3. Compare `go-N...-rolling.csv` vs `ts-N...-rolling.csv` on same plot axes.
4. Document threats: event-loop timers, ms vs µs clocks, same-host loader limits.

---

## Code conventions

- Ignored returns: bare `fn()` when dropping all values.
- `_ = fn()` only for best-effort close/shutdown paths.
- Hot path (dial, read, write, parse): check `err`.

---

## Todo checklist

- [x] `cmd/load` — print GameOver winner line on stdout
- [x] `scripts/load-parse-results` → `{run_id}-rolling.csv` + `runs.csv`
- [x] `scripts/benchmark.py` — rolling plots (`window_end` + `active_games`)
- [x] End-to-end pipeline smoke
- [ ] Go N ladder until saturation (3× per step)
- [ ] Optional: summary plot from `runs.csv`
- [ ] Optional: CPU/RAM (`measure` / pidstat)
- [ ] Optional: perf + FlameGraph
- [ ] TypeScript server + Go vs TS overlay plots

---

## Out of scope / not doing

- tyrone `test_client` / chat latency pipeline
- `games.csv` (`.res` is raw per-game archive)
- Auto-sweep runner (until manual runs are painful)
- Old PLAN step table / pprof checklist
- Server log capture (watch server in terminal only)

---

## New chat starter

```
Benchmark work in go-runs-rest-walk. Load .res → rolling CSV + runs.csv → plots. @.cursor/PLAN.md
Reference ../tyrone-biggums for game_player + go.N.res.csv shape.
Parser: scripts/load-parse-results (Go). Plots: scripts/benchmark.py (Python).
Next: Go N saturation sweep, then TypeScript server.
```
