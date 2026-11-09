FROM gcr.io/distroless/static-debian11:latest@sha256:1dbe426d60caed5d19597532a2d74c8056cd7b1674042b88f7328690b5ead8ed
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/checkpointz* /checkpointz
ENTRYPOINT ["/checkpointz"]
