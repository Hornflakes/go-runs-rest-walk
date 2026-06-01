package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	runFileRx   = regexp.MustCompile(`^(.+)-N(\d+)-G(\d+)-run(\d+)-load\.res$`)
	loadStartRx = regexp.MustCompile(
		`load \| url=(\S+) connections=(\d+) games=(\d+) stagger=(\S+) fire=(\S+)`,
	)
	loadFinishRx = regexp.MustCompile(
		`load finished \| games_over=(\d+) games_failed=(\d+) clients_started=(\d+)/(\d+) clients_done=(\d+)/(\d+)`,
	)
	ansiRx   = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	winnerRx = regexp.MustCompile(
		`^winner=(\d+) histogram=([\d,]+) active_games=(\d+)$`,
	)
)

type loadMeta struct {
	url         string
	connections int
	games       int
	stagger     string
	fire        string
}

type loadFinish struct {
	gamesOver      uint64
	gamesFailed    uint64
	clientsStarted uint64
	connections    uint64
	clientsDone    uint64
}

type runName struct {
	language  string
	runNumber int
}

type game struct {
	activeGames uint64
	buckets     [8]uint64
}

func runIdFromInPath(inPath string) string {
	base := filepath.Base(inPath)
	if runId, ok := strings.CutSuffix(base, "-load.res"); ok {
		return runId
	}
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func parseRunName(base string) runName {
	ms := runFileRx.FindStringSubmatch(base)
	if ms == nil {
		return runName{}
	}

	runNum, _ := strconv.Atoi(ms[4])

	return runName{
		language:  ms[1],
		runNumber: runNum,
	}
}

func stripANSI(s string) string {
	return ansiRx.ReplaceAllString(s, "")
}

func parseBuckets(raw string) ([8]uint64, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 8 {
		return [8]uint64{}, fmt.Errorf("got %d, want 8 buckets", len(parts))
	}

	var buckets [8]uint64
	for i, p := range parts {
		n, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return [8]uint64{}, err
		}
		buckets[i] = n
	}
	return buckets, nil
}

func parseWinnerLine(line string) (game, bool) {
	ms := winnerRx.FindStringSubmatch(strings.TrimSpace(line))
	if ms == nil {
		return game{}, false
	}

	buckets, err := parseBuckets(ms[2])
	if err != nil {
		return game{}, false
	}

	activeGames, err := strconv.ParseUint(ms[3], 10, 64)
	if err != nil {
		return game{}, false
	}

	return game{activeGames, buckets}, true
}

func readLoadFile(path string) ([]game, loadMeta, loadFinish, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, loadMeta{}, loadFinish{}, err
	}
	defer f.Close()

	var games []game
	var meta loadMeta
	var finish loadFinish
	var haveFinish bool

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := stripANSI(sc.Text())

		if ms := loadStartRx.FindStringSubmatch(line); ms != nil {
			meta.url = ms[1]
			meta.connections, _ = strconv.Atoi(ms[2])
			meta.games, _ = strconv.Atoi(ms[3])
			meta.stagger = ms[4]
			meta.fire = ms[5]
			continue
		}

		if ms := loadFinishRx.FindStringSubmatch(line); ms != nil {
			finish.gamesOver, _ = strconv.ParseUint(ms[1], 10, 64)
			finish.gamesFailed, _ = strconv.ParseUint(ms[2], 10, 64)
			finish.clientsStarted, _ = strconv.ParseUint(ms[3], 10, 64)
			finish.clientsDone, _ = strconv.ParseUint(ms[5], 10, 64)
			haveFinish = true
			continue
		}

		if g, ok := parseWinnerLine(line); ok {
			games = append(games, g)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, loadMeta{}, loadFinish{}, err
	}
	if !haveFinish {
		log.Printf("warning: no load finished line in %s", path)
	}
	return games, meta, finish, nil
}

func bucket0Ratio(games []game) float64 {
	var b0, total uint64
	for _, g := range games {
		for i, n := range g.buckets {
			total += n
			if i == 0 {
				b0 += n
			}
		}
	}

	if total == 0 {
		return 0
	}
	return float64(b0) / float64(total)
}

func appendRunsCSV(path, runId string, name runName, meta loadMeta, finish loadFinish, games []game) error {
	needHeader := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		needHeader = true
	} else if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if needHeader {
		if err := w.Write([]string{
			"run_id", "language", "connections", "games_per_conn", "run_number",
			"url", "stagger", "fire",
			"games_parsed", "bucket0_ratio_overall",
			"games_over", "games_failed", "clients_started", "clients_done",
		}); err != nil {
			return err
		}
	}

	language := name.language
	connections := meta.connections
	gamesPerConn := meta.games
	runNumber := strconv.Itoa(name.runNumber)

	if err := w.Write([]string{
		runId,
		language,
		strconv.Itoa(connections),
		strconv.Itoa(gamesPerConn),
		runNumber,
		meta.url,
		meta.stagger,
		meta.fire,
		strconv.Itoa(len(games)),
		fmt.Sprintf("%.2f", bucket0Ratio(games)),
		strconv.FormatUint(finish.gamesOver, 10),
		strconv.FormatUint(finish.gamesFailed, 10),
		strconv.FormatUint(finish.clientsStarted, 10),
		strconv.FormatUint(finish.clientsDone, 10),
	}); err != nil {
		return err
	}
	return w.Error()
}

func rollingPath(outPath, runId string) string {
	dir := filepath.Dir(outPath)
	return filepath.Join(dir, runId+"-rolling.csv")
}

func meanActiveGames(games []game) float64 {
	var sum uint64
	for _, g := range games {
		sum += g.activeGames
	}
	return float64(sum) / float64(len(games))
}

func writeRolling(outPath string, games []game, window int) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"window_end", "active_games", "bucket0_ratio"}); err != nil {
		return err
	}

	for end := window; end <= len(games); end += window {
		chunk := games[end-window : end]
		if err := w.Write([]string{
			strconv.Itoa(end),
			fmt.Sprintf("%.2f", meanActiveGames(chunk)),
			fmt.Sprintf("%.2f", bucket0Ratio(chunk)),
		}); err != nil {
			return err
		}
	}

	return w.Error()
}

func main() {
	window := flag.Int("window", 100, "games per rolling window")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("usage: go run ./load-parse-results [-window 100] <file.res>")
	}

	inPath := flag.Arg(0)
	runId := runIdFromInPath(inPath)
	name := parseRunName(filepath.Base(inPath))

	games, meta, finish, err := readLoadFile(inPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(games) == 0 {
		log.Fatal("no winner lines found")
	}

	runsPath := filepath.Join(filepath.Dir(inPath), "runs.csv")
	if err := appendRunsCSV(runsPath, runId, name, meta, finish, games); err != nil {
		log.Fatal(err)
	}

	if len(games) < *window {
		log.Fatalf("got %d games, want at least %d for window size (use -window %d or run more games)",
			len(games), *window, len(games))
	}

	rollingOutPath := rollingPath(inPath, runId)
	if err := writeRolling(rollingOutPath, games, *window); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("parsed %d games => wrote %s, appended %s\n", len(games), rollingOutPath, runsPath)
}
