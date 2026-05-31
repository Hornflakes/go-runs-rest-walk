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
- Disconnect + context (ready phase), `GameQueue.Stop()`, safe sleep

---

## Wire protocol (actual)

JSON: `{"type": n, "msg"?: "..."}`

| type | Name     | Notes                                             |
| ---- | -------- | ------------------------------------------------- |
| 0    | Hello    | Server → client                                   |
| 1    | Ready    | Handshake echo                                    |
| 2    | Shoot    | Client → server                                   |
| 3    | GameOver | Winner: `winner(activeGames)=bucket0,bucket1,...` |

Constants: see `internal/gameloop/spec.go` (tick 16 ms, fire rates, etc.).

---

## Benchmark philosophy (tyrone shooter model)

- **Load client** is the benchmark harness (like tyrone `game_player`).
- **Histogram data** comes from **load stdout** (GameOver `winner(...)` lines), teed to `.res` files.
- **Server:** run in a terminal and watch live — **no server log capture**.
- **Same `cmd/load`** against Go and (later) TypeScript servers for fair comparison.

---

## Benchmark workflow

### 1. Run capture

```bash
# Terminal 1 — server (watch in terminal, no tee)
go run ./go/cmd/server

# Terminal 2 — load (tee stdout for scripts)
go run ./go/cmd/load -connections N -stagger 5 -fire 200 -games 0 2>&1 | tee load-go-N-run1.res
```

Per run document: `run_id`, language, connections, stagger, steady duration (~5 min after `clients_started=N/N`), OS, CPU/RAM, git commit.

Repeat **≥3 runs** per N. N ladder e.g. 500 → 1k → 2k → 5k → 8k → 10k (stop when `games_failed` or tick quality collapses).

### 2. Load changes (todo)

- On every GameOver, **always** print `msg` to stdout (tyrone-style `winner(N)=...` line)
- Keep existing heartbeats + finish milestone for `runs.csv` client fields

### 3. Scripts (todo)

**`scripts/parse-load-res`** — stdin = load `.res` file

- Parse lines starting with `winner(`
- Support format: `winner(activeGames)=b0,b1,...,b7` (and optional tyrone `___` for reference files)

**Output: `games.csv`** (one row per game)

```text
run_id,game_idx,active_games,w0,w1,w2,w3,w4,w5,w6,w7
```

**Output: `rolling.csv`** (every 100 games, tyrone `go.N.res.csv` shape)

```text
run_id,window_end,active_games,pct_bucket0
```

- `pct_bucket0 = sum(w0) / sum(w0..w7) × 100`
- `active_games` = mean of `activeGames` from winner lines in window

**Output: append row to `runs.csv`**

```text
run_id,language,connections,run_number,stagger_ms,steady_s,
games_parsed,pct_bucket0_overall,games_over,games_failed,clients_started,notes
```

- Histogram fields from parsed games; `games_*` / `clients_*` from load finish/heartbeat lines in same `.res` file

### 4. Plots (todo)

**`scripts/plot_benchmarks.py`**

- **Rolling chart:** X = `active_games`, Y = `pct_bucket0`, one line per `run_id` → main tyrone-style curve
- **Summary chart:** X = `connections`, Y = median `pct_bucket0_overall` (3 runs) → saturation comparison
- Later: overlay **Go vs TypeScript** on same axes

### 5. CPU/RAM (optional, tyrone `measure`)

During steady state, sample server + load PIDs (`pidstat` or copy tyrone `measure`).

Separate from histogram pipeline; supports “loader vs server CPU” narrative.

### 6. Profiling (optional)

- `perf` + FlameGraph on server PID at low / mid / saturated N (not pprof)
- Linux/WSL for published numbers

---

## Go vs TypeScript (later)

1. Port TS server to same spec + wire format.
2. Same load flags, same N ladder, same scripts.
3. `runs.csv` column `language=go|typescript` → comparison plots.
4. Document threats: event-loop timers, ms vs µs clocks, same-host loader limits.

---

## Code conventions

- Ignored returns: bare `fn()` when dropping all values.
- `_ = fn()` only for best-effort close/shutdown paths.
- Hot path (dial, read, write, parse): check `err`.

---

## Todo checklist

- [ ] `cmd/load` — always print GameOver winner line on stdout
- [ ] `scripts/parse-load-res` → games.csv + rolling.csv + runs row
- [ ] `scripts/plot_benchmarks.py`
- [ ] First full run (e.g. N=5000, 3×) + one rolling plot
- [ ] Load sweep + runs.csv
- [ ] perf + FlameGraph (optional)
- [ ] TypeScript server + Go vs TS plots

---

## Out of scope / not doing

- tyrone `test_client` / chat latency pipeline
- Auto-sweep runner (until manual runs are painful)
- Old PLAN step table / pprof checklist
- Server log capture (watch server in terminal only)

---

## New chat starter

```
Benchmark work in go-runs-rest-walk. Load .res → CSV → plots. @.cursor/PLAN.md
Reference ../tyrone-biggums for game_player + process-game-server-results shape.
```
