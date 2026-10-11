FROM golang:1.26-bookworm@sha256:13e7249b4618c115a175ea2627213131855233ecf465328cac30a0f754beb985 AS builder
WORKDIR /src
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app .

FROM ubuntu:24.04@sha256:786a8b558f7be160c6c8c4a54f9a57274f3b4fb1491cf65146521ae77ff1dc54
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
  libssl-dev \
  ca-certificates \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --uid 10001 --no-create-home --user-group checkpointz
COPY --from=builder /bin/app /checkpointz
USER checkpointz
EXPOSE 5555
ENTRYPOINT ["/checkpointz"]
