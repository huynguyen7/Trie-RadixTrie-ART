import argparse
import json
import statistics
import subprocess
import tempfile
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt


ROOT = Path(__file__).resolve().parent
OUTPUT = ROOT / "bench-results"
REPEATS = 3
BENCH_TIME = "100ms"
TREES = {"trie": "Trie", "radix": "Radix trie", "art": "ART"}
DATASETS = {
    "shared": ("Same first 16 letters", "tab:blue", "-", "o"),
    "scattered": ("Random first 16 letters", "tab:orange", "--", "s"),
}


def plot_lines(ax, results, metric, title, unit, divisor=1):
    names = list(TREES)
    highest = 0

    for dataset, (label, color, style, marker) in DATASETS.items():
        values = []
        for i, name in enumerate(names):
            samples = [row[metric] / divisor for row in results
                       if row["implementation"] == name and row["dataset"] == dataset]
            value = statistics.median(samples)
            values.append(value)
            highest = max(highest, value)
            offset = -8 if dataset == "shared" else 8
            align = "right" if dataset == "shared" else "left"
            ax.annotate(f"{value:,.2f}", (i, value), xytext=(offset, 10),
                        textcoords="offset points", ha=align, color=color)
        ax.plot(range(len(names)), values, color=color, linestyle=style,
                marker=marker, linewidth=2, markersize=6, label=label)

    ax.set_xticks(range(len(names)), [TREES[name] for name in names])
    ax.set(title=title, xlabel="Tree implementation", ylabel=unit,
           xlim=(-0.5, len(names) - 0.5),
           ylim=(0, highest * 1.3 if highest else 1))
    ax.grid(axis="y", alpha=0.2)
    ax.set_axisbelow(True)


def draw_charts(results):
    keys = results[0]["keys"]
    panels = [
        ("lookup_ns", "Find an existing key", "Average nanoseconds per key", 1),
        ("insert_ns", "Insert a new key", "Average nanoseconds per key", 1),
        ("delete_ns", "Delete an existing key", "Average nanoseconds per key", 1),
        ("walkprefix_ns", f"Walk all {keys:,} keys", "Milliseconds per full walk", 1e6),
    ]

    fig, axes = plt.subplots(2, 2, figsize=(12, 8))
    for ax, (metric, title, unit, divisor) in zip(axes.flat, panels):
        plot_lines(ax, results, metric, title, unit, divisor)
    handles, labels = axes[0, 0].get_legend_handles_labels()
    fig.legend(handles, labels, loc="lower center", ncol=2)
    fig.text(0.5, 0.06, "Each key has 16 prefix letters and a unique six-letter ending. Lower values are better.",
             ha="center", fontsize=9)
    fig.suptitle(f"{keys:,} keys\nFixed length of 22 bytes per key")
    fig.tight_layout(rect=(0, 0.09, 1, 0.95))
    fig.savefig(OUTPUT / "operations.png", dpi=160)
    plt.close(fig)

    fig, ax = plt.subplots(figsize=(8, 5))
    plot_lines(ax, results, "bytes_per_key", f"Tree memory for {keys:,} keys\nFixed length of 22 bytes per key",
                "Bytes per stored key")
    ax.legend(loc="upper left")
    fig.text(0.5, 0.02, "Approximate extra heap after building the tree. Original key strings are excluded.",
             ha="center", fontsize=9)
    fig.tight_layout(rect=(0, 0.05, 1, 1))
    fig.savefig(OUTPUT / "memory.png", dpi=160)
    plt.close(fig)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--sizes", type=int, default=10000, help="Number of keys (one size)")
    args = parser.parse_args()
    if args.sizes < 1:
        parser.error("sizes must be positive")

    tests = subprocess.run(["go", "test", "./..."], cwd=ROOT,
                           capture_output=True, text=True)
    if tests.returncode:
        raise SystemExit(tests.stdout + tests.stderr)

    results = []
    with tempfile.TemporaryDirectory() as temporary:
        binary = Path(temporary) / "bench"
        subprocess.run(["go", "build", "-o", str(binary), "./bench"],
                       cwd=ROOT, check=True)
        for _ in range(REPEATS):
            text = subprocess.check_output(
                [str(binary), "-keys", str(args.sizes),
                 "-test.benchtime=" + BENCH_TIME], cwd=ROOT, text=True)
            results.extend(json.loads(text))

    OUTPUT.mkdir(exist_ok=True)
    draw_charts(results)
    print(f"Charts saved to {OUTPUT}")


if __name__ == "__main__":
    main()

