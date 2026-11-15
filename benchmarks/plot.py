import math
import sys
from pathlib import Path

import matplotlib

matplotlib.use("svg")
import matplotlib.pyplot as plt

LANGUAGES = ["C", "Go", "Python", "Ricol"]
SERIES_COLORS = ["#2a78d6", "#eb6834", "#1baf7a", "#eda100", "#e87ba4", "#008300", "#6250d6", "#e34948"]
SURFACE = "#fcfcfb"
TEXT_PRIMARY = "#0b0b0b"
TEXT_SECONDARY = "#52514e"
MUTED = "#898781"
GRIDLINE = "#e1e0d9"
BASELINE_VERSION = "baseline"
MILLISECONDS_PER_UNIT = {"s": 1000, "ms": 1, "µs": 0.001}
VERSION_LABELS = {
    "first_version": "Primera versión",
    "baseline": "Base",
    "constant_folding": "Constant folding",
    "dead_code_elimination": "Dead code elim.",
    "lazy_environments": "Maps perezosos",
    "node_distances": "Distancia en el nodo",
    "return_signal": "Return directo",
    "no_grouping": "Sin agrupaciones",
}

plt.rcParams.update({
    "svg.fonttype": "none",
    "svg.hashsalt": "ricol",
    "font.size": 10,
    "figure.facecolor": SURFACE,
    "axes.facecolor": SURFACE,
    "axes.edgecolor": GRIDLINE,
    "axes.labelcolor": TEXT_SECONDARY,
    "axes.titlecolor": TEXT_PRIMARY,
    "axes.titlesize": 12,
    "axes.titlelocation": "left",
    "axes.spines.top": False,
    "axes.spines.right": False,
    "axes.spines.left": False,
    "axes.grid": True,
    "axes.grid.axis": "y",
    "axes.axisbelow": True,
    "grid.color": GRIDLINE,
    "grid.linewidth": 0.8,
    "xtick.color": TEXT_SECONDARY,
    "ytick.color": TEXT_SECONDARY,
    "ytick.left": False,
    "legend.frameon": False,
    "legend.labelcolor": TEXT_PRIMARY,
})


def version_label(version):
    return VERSION_LABELS.get(version, version)


def results_path(resultsdir, version):
    return resultsdir / f"results_{version}.md"


def load_version(path):
    results = {}
    benchmark = None
    milliseconds_per_unit = None
    for line in path.read_text().splitlines():
        if line.startswith("## "):
            benchmark = line[3:].strip()
            milliseconds_per_unit = None
        elif benchmark and line.startswith("| Command |"):
            unit = line.split("Mean [")[1].split("]")[0]
            milliseconds_per_unit = MILLISECONDS_PER_UNIT[unit]
        elif milliseconds_per_unit and line.startswith("| `"):
            cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
            command = cells[0].strip("`")
            mean = float(cells[1].split("±")[0])
            results.setdefault(benchmark, {})[command] = mean * milliseconds_per_unit
    return results


def format_milliseconds(milliseconds):
    if milliseconds >= 100:
        return f"{milliseconds:.0f}"
    if milliseconds >= 10:
        return f"{milliseconds:.1f}"
    return f"{milliseconds:.2f}"


def save(figure, path):
    figure.tight_layout()
    figure.savefig(path, format="svg", metadata={"Date": None})
    plt.close(figure)


def plot_languages(version, results, chartsdir):
    benchmarks = list(results)
    slot_width = 0.8 / len(LANGUAGES)
    figure, axes = plt.subplots(figsize=(10, 4.8))
    for index, language in enumerate(LANGUAGES):
        positions = [position + (index - (len(LANGUAGES) - 1) / 2) * slot_width for position in range(len(benchmarks))]
        values = [results[benchmark].get(language, math.nan) for benchmark in benchmarks]
        bars = axes.bar(positions, values, width=slot_width * 0.9, color=SERIES_COLORS[index], label=language)
        if language == "Ricol":
            axes.bar_label(bars, labels=[format_milliseconds(value) for value in values],
                           padding=2, fontsize=8, color=TEXT_SECONDARY)
    axes.set_yscale("log")
    axes.set_xticks(range(len(benchmarks)), benchmarks)
    axes.set_ylabel("Tiempo medio [ms], escala logarítmica")
    axes.set_title(f"Tiempo por benchmark y lenguaje: {version_label(version)}")
    axes.legend(ncols=len(LANGUAGES), loc="upper left", bbox_to_anchor=(0, -0.08))
    save(figure, chartsdir / f"{version}-languages.svg")


def plot_before_after(previous, current, results_by_version, chartsdir):
    previous_results = results_by_version[previous]
    current_results = results_by_version[current]
    benchmarks = [benchmark for benchmark in current_results if "Ricol" in previous_results.get(benchmark, {})]
    previous_values = [previous_results[benchmark]["Ricol"] for benchmark in benchmarks]
    current_values = [current_results[benchmark]["Ricol"] for benchmark in benchmarks]
    slot_width = 0.38
    figure, axes = plt.subplots(figsize=(10, 4.8))
    axes.bar([position - slot_width / 2 for position in range(len(benchmarks))], previous_values,
             width=slot_width * 0.94, color=SERIES_COLORS[0], label=version_label(previous))
    axes.bar([position + slot_width / 2 for position in range(len(benchmarks))], current_values,
             width=slot_width * 0.94, color=SERIES_COLORS[1], label=version_label(current))
    for position, (before, after) in enumerate(zip(previous_values, current_values)):
        change = (after - before) / before * 100
        axes.annotate(f"{change:+.1f}%", (position, max(before, after)), xytext=(0, 4),
                      textcoords="offset points", ha="center", fontsize=9, color=TEXT_PRIMARY)
    axes.set_xticks(range(len(benchmarks)), benchmarks)
    axes.set_ylabel("Tiempo medio de Ricol [ms]")
    axes.set_ylim(0, max(previous_values + current_values) * 1.12)
    axes.set_title(f"Ricol antes y después: {version_label(previous)} → {version_label(current)}")
    axes.legend(ncols=2, loc="upper left", bbox_to_anchor=(0, -0.08))
    save(figure, chartsdir / f"{current}-before-after.svg")


def plot_evolution(versions, results_by_version, chartsdir):
    baseline = results_by_version[BASELINE_VERSION]
    benchmarks = [benchmark for benchmark in baseline if "Ricol" in baseline[benchmark]]
    figure, axes = plt.subplots(figsize=(10, 5.2))
    axes.axhline(100, color=MUTED, linewidth=1, linestyle="--")
    for index, benchmark in enumerate(benchmarks):
        relative_times = []
        for version in versions:
            measured = results_by_version[version].get(benchmark, {}).get("Ricol")
            relative_times.append(measured / baseline[benchmark]["Ricol"] * 100 if measured else math.nan)
        color = SERIES_COLORS[index % len(SERIES_COLORS)]
        axes.plot(range(len(versions)), relative_times, color=color, linewidth=2,
                  marker="o", markersize=6, markeredgecolor=SURFACE, markeredgewidth=1.5, label=benchmark)
    axes.set_xticks(range(len(versions)), [version_label(version) for version in versions],
                    rotation=20, ha="right")
    axes.set_ylabel("Tiempo de Ricol, % de la versión base")
    axes.set_ylim(bottom=0)
    axes.set_title("Evolución del tiempo de Ricol por benchmark")
    axes.legend(loc="upper left", bbox_to_anchor=(1.01, 1))
    save(figure, chartsdir / "evolution.svg")


def main():
    if len(sys.argv) < 2:
        print("Usage: plot.py <results directory> [version ...]", file=sys.stderr)
        sys.exit(1)
    resultsdir = Path(sys.argv[1])
    chartsdir = resultsdir / "charts"
    chartsdir.mkdir(exist_ok=True)
    versions = [version for version in sys.argv[2:] if results_path(resultsdir, version).is_file()]
    results_by_version = {version: load_version(results_path(resultsdir, version)) for version in versions}

    for index, version in enumerate(versions):
        plot_languages(version, results_by_version[version], chartsdir)
        if index > 0:
            plot_before_after(versions[index - 1], version, results_by_version, chartsdir)
    if BASELINE_VERSION in results_by_version:
        plot_evolution(versions, results_by_version, chartsdir)
    print(f"Charts saved to {chartsdir}")


if __name__ == "__main__":
    main()
