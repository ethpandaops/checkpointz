FROM debian:bookworm-slim@sha256:96e378d7e6531ac9a15ad505478fcc2e69f371b10f5cdf87857c4b8188404716
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
  libssl-dev \
  ca-certificates \
  curl \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --uid 10001 --no-create-home --user-group checkpointz
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/checkpointz* /checkpointz
USER checkpointz
ENTRYPOINT ["/checkpointz"]
