FROM golang:1.27-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 hyperfine \
    && rm -rf /var/lib/apt/lists/*

ENV GOCACHE=/tmp/go-cache

WORKDIR /ricol
COPY src/ src/
RUN cd src && go mod download
COPY benchmarks/ benchmarks/

ENTRYPOINT ["./benchmarks/run.sh"]
