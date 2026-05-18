# MSc game server — implementation plan

## Repos

| Role                      | Path                                                                                                |
| ------------------------- | --------------------------------------------------------------------------------------------------- |
| **Implement (this repo)** | `go-runs-rest-walk` — module `github.com/hornflakes/go-runs-rest-walk`, Go 1.26.3                   |
| **Reference (peek only)** | `../tyrone-biggums` — [ThePrimeagen/tyrone-biggums](https://github.com/ThePrimeagen/tyrone-biggums) |
| **Workspace**             | `go-runs-rest-walk.code-workspace` (both folders)                                                   |

**Attribution:** Implementation inspired by tyrone-biggums; write your own code; cite in thesis.

---

## Committed scope

1. **Go server** — steps 0–11 + chosen polish
2. **Go clients** — `cmd/bot` + `cmd/load`
3. **Thorough benchmarking** — see checklist below

**Optional later:** TypeScript/Java, cloud vertical scaling/cost, Rust.

---

## Architecture

- WebSocket on **port 37373**, path `/`
- **Pair** two clients → **ReadyUp** (30s timeout) → **Play**
- **~60 Hz loop:** flush `Fire` → move bullets → AABB collide → sleep ~16 ms
- **GameOver** with frame histogram + `activeGames`
- Clients only send `Fire` (no position sync)

### Wire protocol

JSON: `{"type": n, "msg"?: "..."}`

| type | Name     | Notes                                               |
| ---- | -------- | --------------------------------------------------- |
| 0    | ReadyUp  | Server asks; client echoes                          |
| 1    | Play     | Simulation started                                  |
| 2    | Fire     | Client → server                                     |
| 3    | GameOver | Winner: `winner(activeGames)___bucket0,bucket1,...` |

### Goroutines (Go)

- **Main:** HTTP listen
- **Pairing consumer:** `for pair := range server.Out { game.Run() }`
- **Per game:** ready + tick loop + `GameQueue`
- **Per socket:** read goroutine + write goroutine

### Player 2

Second connector → player index 1, **300 ms** fire vs **180 ms** — handicapped, **not** auto-loss. Winner = first bullet hit.

### Why ~16 ms per tick?

60 ticks/s ≈ 16.67 ms/frame. Reference sleeps **16_000 µs**. Histograms measure **start-to-start** gap; buckets at 17 ms, 19 ms, … = server lag under load.

---

## Implementation steps (0–12)

**Workflow:** open reference file(s) → code in **this** repo → `go test` / `cmd/bot`.

| Step | You build                        | Verify                  | Reference (`../tyrone-biggums/`)            |
| ---- | -------------------------------- | ----------------------- | ------------------------------------------- |
| 0    | `go.mod`, `gorilla/websocket`    | `go build ./...`        | `go/go.mod`                                 |
| 1    | Protocol + JSON                  | Unit test `{"type":2}`  | `go/pkg/server/message.go`                  |
| 2    | WS upgrade, echo JSON            | `websocat` or early bot | `go/pkg/server/socket.go`                   |
| 3    | Pairing (1st waits, 2nd pairs)   | Log "paired"            | `go/pkg/server/server.go`                   |
| 4    | Read/write goroutines + chans    | Echo via channels       | `go/pkg/server/socket.go`                   |
| 5    | Ready handshake, 30s timeout     | 2 bots                  | `go/pkg/game_loop/wait_for_ready.go`        |
| 6    | Empty tick loop + `Play`         | ~180 ticks in 3s        | `go/pkg/game_loop/game.go`, `game_clock.go` |
| 7    | `GameQueue` + `Flush`            | `go test`               | `game_queue.go`, `game_queue_test.go`       |
| 8    | Players, fire rate, bullets      | Flush → bullets         | `objects.go`, `objects_test.go`             |
| 9    | Move + AABB + `GameOver`         | 2 bots, match ends      | `game.go`, `geometry.go`                    |
| 10   | Histogram + `activeGames`        | Parse winner line       | `go/pkg/stats/stats.go`                     |
| 11   | Wire `main`                      | Full duel               | `go/cmd/server/main.go`                     |
| 12   | Polish + `cmd/load` + benchmarks | Load sweep, `pprof`     | `run-server`, `measure` (optional)          |

**cmd/bot** (start ~step 2–3): on `ReadyUp` → reply `ReadyUp`; on `Play` → `Fire` every **200 ms**; on `GameOver` → log, exit.

**cmd/load** (step 12): `-connections N`, stagger **50 ms**; N connections ≈ **N/2** games.

Reference clients (optional peek): `typescript/src/test/game_player.ts`, `rust/src/bin/game_player.rs`.

---

## Go improvements (pick 3–4)

| Improvement                              | Why                                             |
| ---------------------------------------- | ----------------------------------------------- |
| `context.Context`                        | Cancel ready, queue, readers on end/disconnect  |
| `GameQueue.Stop()` on game end           | Reference queue may outlive match               |
| Safe sleep `max(0, 16ms - elapsed)`      | Avoid negative `time.Sleep`                     |
| Handle `Game.Run()` completion in `main` | Log errors/timeouts                             |
| Written spec constants                   | 100×100 player, 180/300 ms fire, bullet speed 1 |
| `net/http/pprof`                         | CPU/heap under load                             |

---

## Benchmarking checklist (thorough)

**Document once:** CPU, RAM, OS, Go version, `ulimit -n`, server/client flags, git commit.

**Functional:** 2× `cmd/bot` → full match + parsed histogram.

**Load sweep:** e.g. 50, 100, 200, 500, 1000, 2000, 4000 — stop at OS/machine limit.

**Repeat:** ≥3 runs per N; median/spread, not one run.

**Record:** histogram buckets, `activeGames`, CPU/RAM.

**Analysis:** plot connections vs % good ticks (bucket 0); define **saturation point** (e.g. >5% bad ticks).

**Profiling:** `pprof` at low / medium / saturated load.

**Artifacts:** CSV/JSON, plots, `go test ./...`, README reproduce steps.

---

## Reference read order (tyrone-biggums Go)

1. `go/cmd/server/main.go`
2. `go/pkg/server/server.go` + `socket.go`
3. `go/pkg/server/message.go`
4. `go/pkg/game_loop/wait_for_ready.go`
5. `go/pkg/game_loop/game_queue.go`
6. `go/pkg/game_loop/game.go`
7. `go/pkg/stats/stats.go`

Run reference tests: `cd ../tyrone-biggums/go && go test ./...`

---

## Methodological validity (thesis)

- tyrone-biggums is **inspiration**, not ground truth for language rankings.
- Uneven optimization across Go/TS/Rust in that repo; different tick drivers; possible constant mismatches (e.g. Rust player height).
- Your work: **written spec**, reproducible harness, **threats to validity**, separate **effort** vs **performance**.

---

## Optional later

- **Vertical scaling:** bigger VM, same binary, repeat load sweep.
- **Horizontal scaling:** needs redesign (matchmaker, sticky WS) — future work.
- **Lambda:** poor fit for persistent WS + 60 Hz loops — architectural contrast only.
- **TypeScript / Java:** same spec, second implementation.

---

## Todo checklist

- [ ] `docs/SPEC.md` or spec section (protocol, constants, tick order, histogram units)
- [ ] Server steps 0–11
- [ ] `cmd/bot`
- [ ] `cmd/load`
- [ ] Polish (context, sleep, queue stop, pprof)
- [ ] Thorough benchmarks + artifacts
- [ ] (Optional) second language, cloud chapter

---

## New chat

```
Implement in go-runs-rest-walk. Reference ../tyrone-biggums. @.cursor/PLAN.md — start step N.
```
