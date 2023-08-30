FROM debian:bullseye
RUN apt update && \
    apt install -y ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
