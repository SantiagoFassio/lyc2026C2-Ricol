.PHONY: integration bench

BENCH_CPU ?= 2

integration:
	./integration/run.sh

bench:
	docker build -f benchmarks/Dockerfile -t ricol-bench .
	mkdir -p benchmarks/results
	docker run --rm \
		--cpuset-cpus=$(BENCH_CPU) \
		--user "$$(id -u):$$(id -g)" \
		-v "$(CURDIR)/benchmarks/results:/ricol/benchmarks/results" \
		ricol-bench
