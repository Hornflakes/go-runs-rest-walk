import argparse
import csv
import re
import statistics
from pathlib import Path

import matplotlib.pyplot as plt

RUN_STEM_RX = re.compile(r"^(?P<group>.+)-run(?P<run>\d+)$")
PLATFORM_COLORS = {"go": "#1ba1e2", "ts": "#008a00", "java": "#e51400"}
FALLBACK_COLORS = ["#9467bd", "#8c564b", "#e377c2", "#7f7f7f", "#bcbd22", "#17becf"]

BIN_WIDTH = 100.0  # active_games bin for the median curve
SAMPLE_INTERVAL = 5.0  # seconds between measure.sh samples (match measure.sh)

LABELS = {
    "ro": {
        "active_title": "Calitatea tick-urilor în funcție de sarcină (mediană)",
        "window_title": "Calitatea tick-urilor pe parcursul rulării (mediană)",
        "cpu_title": "Utilizarea CPU a serverului sub sarcină (mediană)",
        "mem_title": "Memoria serverului sub sarcină (mediană)",
        "active_x": "jocuri active (medie glisantă)",
        "window_x": "jocuri finalizate (sfârșit de fereastră)",
        "bucket0_y": "proporție bucket 0",
        "cpu_y": "CPU (%)",
        "mem_y": "RSS (MB)",
        "elapsed_x": "timp scurs (s)",
    },
    "en": {
        "active_title": "Tick quality vs load (median)",
        "window_title": "Tick quality over run progress (median)",
        "cpu_title": "Server CPU during load (median)",
        "mem_title": "Server memory during load (median)",
        "active_x": "active games (rolling mean)",
        "window_x": "games completed (window end)",
        "bucket0_y": "bucket 0 ratio",
        "cpu_y": "CPU %",
        "mem_y": "RSS (MB)",
        "elapsed_x": "elapsed (s)",
    },
}


def parse_run_stem(stem):
    m = RUN_STEM_RX.match(stem)
    return (m.group("group"), int(m.group("run"))) if m else None


def stem_from_path(path, suffix):
    name = path.name
    return name[: -len(suffix)] if name.endswith(suffix) else path.stem


def to_float(s):
    try:
        return float(str(s).strip().replace(",", "."))
    except (ValueError, AttributeError):
        return None


def fmt(v, decimals=2):
    return "n/a" if v is None else f"{v:.{decimals}f}"


def load_rolling(path):
    window_end, active_games, bucket0 = [], [], []
    with path.open(newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            we, ag, b0 = (
                to_float(row.get("window_end")),
                to_float(row.get("active_games")),
                to_float(row.get("bucket0_ratio")),
            )
            if None in (we, ag, b0):
                continue
            window_end.append(we)
            active_games.append(ag)
            bucket0.append(b0)
    return window_end, active_games, bucket0


def load_measure(path):
    cpus, mems = [], []
    with path.open(newline="", encoding="utf-8") as f:
        rows = list(csv.DictReader(f))
    i = 0
    while i < len(rows):
        if (
            rows[i].get("Type") == "CPU"
            and i + 1 < len(rows)
            and rows[i + 1].get("Type") == "MEM"
        ):
            cpu = to_float(rows[i].get("Value"))
            mem = to_float(str(rows[i + 1].get("Value", "")).replace("MB", ""))
            if cpu is not None and mem is not None:
                cpus.append(cpu)
                mems.append(mem)
            i += 2
        else:
            i += 1
    elapsed = [SAMPLE_INTERVAL * (j + 1) for j in range(len(cpus))]
    return elapsed, cpus, mems


def group_paths_by_stem(paths, suffix):
    groups = {}
    for path in paths:
        parsed = parse_run_stem(stem_from_path(path, suffix))
        if parsed is None:
            print(f"warning: skipping {path.name} (no -runN- in name)")
            continue
        groups.setdefault(parsed[0], []).append(path)
    for label in groups:
        groups[label].sort(key=lambda p: parse_run_stem(stem_from_path(p, suffix))[1])
    return groups


def load_runs_by_id(results_dir):
    runs_path = results_dir / "runs.csv"
    if not runs_path.is_file():
        return {}
    out = {}
    with runs_path.open(newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            rid = row.get("run_id", "")
            if rid:
                out[rid] = row
    return out


def median_run_ids(groups, runs_by_id):
    ids = {}
    for label, paths in groups.items():
        stems = [stem_from_path(p, "-rolling.csv") for p in paths]
        rows = [runs_by_id[s] for s in stems if s in runs_by_id]
        if rows:
            rows = sorted(
                rows, key=lambda r: to_float(r.get("bucket0_ratio_overall")) or 0.0
            )
            ids[label] = rows[len(rows) // 2].get("run_id", "")
        else:
            ids[label] = stems[len(stems) // 2]
            print(
                f"warning: no runs.csv rows for {label}; plotting middle run ({ids[label]})"
            )
    return ids


def binned_median(xs, ys):
    if not xs:
        return [], []
    bx, by, edge, max_x = [], [], 0.0, max(xs)
    while edge < max_x + BIN_WIDTH:
        in_bin = [ys[i] for i in range(len(xs)) if edge <= xs[i] < edge + BIN_WIDTH]
        if in_bin:
            bx.append(edge + BIN_WIDTH / 2)
            by.append(statistics.median(in_bin))
        edge += BIN_WIDTH
    return bx, by


def median_of(values):
    return statistics.median(values) if values else None


def assign_colors(labels):
    colors, used = {}, set()
    for label in sorted(labels):
        token = label.split("-")[0]
        c = PLATFORM_COLORS.get(token)
        if c is None or c in used:
            free = [x for x in FALLBACK_COLORS if x not in used]
            c = free[0] if free else FALLBACK_COLORS[len(used) % len(FALLBACK_COLORS)]
        used.add(c)
        colors[label] = c
    return colors


def plot_active_games(results_dir, median_ids, colors, out_path, labels):
    fig, ax = plt.subplots(figsize=(10, 6))
    for label in sorted(median_ids):
        path = results_dir / f"{median_ids[label]}-rolling.csv"
        if not path.is_file():
            continue
        _, xs, ys = load_rolling(path)
        if not xs:
            continue
        color = colors[label]
        ax.scatter(xs, ys, s=10, alpha=0.18, color=color)
        bx, by = binned_median(xs, ys)
        ax.plot(
            bx, by, linewidth=2.2, marker="o", markersize=3, color=color, label=label
        )
    ax.set_xlabel(labels["active_x"])
    ax.set_title(labels["active_title"])
    ax.set_ylabel(labels["bucket0_y"])
    ax.set_ylim(0, 1.05)
    ax.grid(True, alpha=0.3)
    ax.legend(loc="best", fontsize=9)
    fig.tight_layout()
    fig.savefig(out_path, dpi=150)
    print(f"wrote {out_path}")


def plot_window_end(results_dir, median_ids, colors, out_path, labels):
    fig, ax = plt.subplots(figsize=(10, 6))
    for label in sorted(median_ids):
        path = results_dir / f"{median_ids[label]}-rolling.csv"
        if not path.is_file():
            continue
        xs, _, ys = load_rolling(path)
        if not xs:
            continue
        ax.plot(xs, ys, linewidth=2.2, color=colors[label], label=label)
    ax.set_xlabel(labels["window_x"])
    ax.set_title(labels["window_title"])
    ax.set_ylabel(labels["bucket0_y"])
    ax.set_ylim(0, 1.05)
    ax.grid(True, alpha=0.3)
    ax.legend(loc="best", fontsize=9)
    fig.tight_layout()
    fig.savefig(out_path, dpi=150)
    print(f"wrote {out_path}")


def plot_measure(results_dir, median_ids, colors, out_path, labels):
    fig, (ax_cpu, ax_mem) = plt.subplots(2, 1, figsize=(10, 8), sharex=True)
    plotted = False
    for label in sorted(median_ids):
        path = results_dir / f"{median_ids[label]}-measure.csv"
        if not path.is_file():
            continue
        elapsed, cpus, mems = load_measure(path)
        if not cpus:
            continue
        color = colors[label]
        ax_cpu.plot(elapsed, cpus, linewidth=2.2, color=color, label=label)
        ax_mem.plot(elapsed, mems, linewidth=2.2, color=color, label=label)
        plotted = True
    if not plotted:
        plt.close(fig)
        print("no *-measure.csv for median runs - skipping measure plot")
        return
    ax_cpu.set_ylabel(labels["cpu_y"])
    ax_cpu.set_title(labels["cpu_title"])
    ax_cpu.grid(True, alpha=0.3)
    ax_cpu.legend(loc="best", fontsize=9)
    ax_mem.set_xlabel(labels["elapsed_x"])
    ax_mem.set_ylabel(labels["mem_y"])
    ax_mem.set_title(labels["mem_title"])
    ax_mem.grid(True, alpha=0.3)
    ax_mem.legend(loc="best", fontsize=9)
    fig.tight_layout()
    fig.savefig(out_path, dpi=150)
    print(f"wrote {out_path}")


def print_summary(groups, runs_by_id, median_ids):
    if not runs_by_id:
        print("no runs.csv - skipping summary")
        return
    print("\nMedian run = middle run by bucket0_ratio_overall (the one plotted).\n")
    for label in sorted(groups):
        stems = [stem_from_path(p, "-rolling.csv") for p in groups[label]]
        rows = sorted(
            (runs_by_id[s] for s in stems if s in runs_by_id),
            key=lambda r: int(r.get("run_number") or 0),
        )
        if not rows:
            continue
        med_id = median_ids.get(label, "")
        print(f"{label}:")
        b0_vals, max_vals = [], []
        for run in rows:
            b0 = to_float(run.get("bucket0_ratio_overall"))
            mx = to_float(run.get("max_active_games"))
            if b0 is not None:
                b0_vals.append(b0)
            if mx is not None:
                max_vals.append(mx)
            marker = "  <- median (plotted)" if run.get("run_id") == med_id else ""
            print(
                f"  run{run.get('run_number', '?')}: bucket0={fmt(b0)} "
                f"max_active={run.get('max_active_games', '?')} "
                f"failed={run.get('games_failed', '0')}{marker}"
            )
        print(
            f"  median: bucket0={fmt(median_of(b0_vals))} max_active={fmt(median_of(max_vals), 0)}\n"
        )


def main():
    p = argparse.ArgumentParser(
        description="Plot benchmark results (median run per label)"
    )
    p.add_argument(
        "results_dir", nargs="?", default="../results", help="dir with result CSVs"
    )
    p.add_argument(
        "--lang", choices=["ro", "en"], default="en", help="figure label language"
    )
    args = p.parse_args()
    labels = LABELS[args.lang]
    results_dir = Path(args.results_dir).resolve()

    groups = group_paths_by_stem(
        sorted(results_dir.glob("*-rolling.csv")), "-rolling.csv"
    )
    if not groups:
        raise SystemExit("no *-rolling.csv files with -runN- in the name found")

    runs_by_id = load_runs_by_id(results_dir)
    median_ids = median_run_ids(groups, runs_by_id)
    colors = assign_colors(median_ids.keys())

    print_summary(groups, runs_by_id, median_ids)
    plot_active_games(
        results_dir,
        median_ids,
        colors,
        results_dir / "rolling-active-games.png",
        labels,
    )
    plot_window_end(
        results_dir, median_ids, colors, results_dir / "rolling-over-time.png", labels
    )
    plot_measure(
        results_dir, median_ids, colors, results_dir / "measure-cpu-mem.png", labels
    )


if __name__ == "__main__":
    main()
