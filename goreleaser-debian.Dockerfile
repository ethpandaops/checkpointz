FROM debian:latest
RUN apt update; \
    apt install -y ca-certificates; \
    rm -rf /var/lib/apt/lists/*
COPY checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
