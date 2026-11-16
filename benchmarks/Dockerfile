FROM golang:1.27-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 python3-matplotlib hyperfine \
    && rm -rf /var/lib/apt/lists/*

ENV GOCACHE=/tmp/go-cache
ENV MPLCONFIGDIR=/tmp/matplotlib

WORKDIR /ricol
COPY --from=ricol-src . src/
RUN cd src && go mod download
COPY benchmarks/ benchmarks/

ENTRYPOINT ["./benchmarks/run.sh"]
