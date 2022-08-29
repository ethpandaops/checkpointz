FROM gcr.io/distroless/static-debian11:latest
COPY checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
