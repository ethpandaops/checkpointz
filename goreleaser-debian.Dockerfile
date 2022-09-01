FROM debian:latest
COPY checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
