FROM gcr.io/distroless/static-debian11:latest
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
