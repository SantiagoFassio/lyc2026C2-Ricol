.PHONY: integration bench bench-charts

BENCH_CPU ?= 2
RICOL_SRC ?= src
BENCH_VERSIONS := first_version baseline constant_folding dead_code_elimination \
	lazy_environments node_distances return_signal no_grouping

integration:
	./integration/run.sh

bench:
	docker build -f benchmarks/Dockerfile --build-context ricol-src=$(RICOL_SRC) -t ricol-bench .
	mkdir -p benchmarks/results
	docker run --rm \
		--cpuset-cpus=$(BENCH_CPU) \
		--user "$$(id -u):$$(id -g)" \
		-v "$(CURDIR)/benchmarks/results:/ricol/benchmarks/results" \
		ricol-bench

bench-charts:
	docker build -f benchmarks/Dockerfile --build-context ricol-src=$(RICOL_SRC) -t ricol-bench .
	docker run --rm \
		--user "$$(id -u):$$(id -g)" \
		-v "$(CURDIR)/benchmarks/results:/ricol/benchmarks/results" \
		--entrypoint python3 \
		ricol-bench benchmarks/plot.py benchmarks/results $(BENCH_VERSIONS)
