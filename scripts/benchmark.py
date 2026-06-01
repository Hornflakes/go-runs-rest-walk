import argparse
import csv
from pathlib import Path

import matplotlib.pyplot as plt

X_CHOICES = ("window_end", "active_games")


def run_id_from_rolling_path(path: Path) -> str:
    name = path.name
    if name.endswith("-rolling.csv"):
        return name[: -len("-rolling.csv")]
    return path.stem


def load_rolling(path: Path, x_key: str) -> tuple[list[float], list[float]]:
    xs: list[float] = []
    ys: list[float] = []

    with path.open(newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            xs.append(float(row[x_key]))
            ys.append(float(row["bucket0_ratio"]))

    return xs, ys


def plot_rollings(paths: list[Path], out_path: Path, x_key: str) -> None:
    if not paths:
        raise SystemExit("no *-rolling.csv files found")

    fig, ax = plt.subplots(figsize=(10, 6))

    for path in sorted(paths):
        xs, ys = load_rolling(path, x_key)
        if not xs:
            continue
        label = run_id_from_rolling_path(path)
        ax.plot(xs, ys, marker="o", markersize=4, linewidth=1.5, label=label)

    if x_key == "window_end":
        ax.set_xlabel("games completed (window end)")
        ax.set_title("Tick quality over run progress")
    else:
        ax.set_xlabel("active games (rolling mean)")
        ax.set_title("Tick quality vs load")

    ax.set_ylabel("bucket 0 ratio")
    ax.set_ylim(0, 1.05)
    ax.grid(True, alpha=0.3)
    ax.legend(loc="best", fontsize=8)
    fig.tight_layout()
    fig.savefig(out_path, dpi=150)

    print(f"wrote {out_path}")


def main() -> None:
    p = argparse.ArgumentParser(description="Plot *-rolling.csv benchmark files")
    p.add_argument(
        "results_dir",
        nargs="?",
        default="../results",
        help="directory containing *-rolling.csv (default: ../results)",
    )
    p.add_argument(
        "-o",
        "--output",
        default="",
        help="output PNG path (default: <results_dir>/rolling.png or rolling-load.png)",
    )
    p.add_argument(
        "--x",
        choices=X_CHOICES,
        default="window_end",
        help="x-axis column (default: window_end)",
    )
    args = p.parse_args()

    results_dir = Path(args.results_dir).resolve()
    paths = list[Path](results_dir.glob("*-rolling.csv"))

    if args.output:
        out = Path(args.output).resolve()
    elif args.x == "active_games":
        out = results_dir / "rolling-active-games.png"
    else:
        out = results_dir / "rolling-over-time.png"

    plot_rollings(paths, out, args.x)


if __name__ == "__main__":
    main()
