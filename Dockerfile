FROM golang:1.18-alpine as builder
RUN apk add --no-cache gcc musl-dev linux-headers git
WORKDIR /src
COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN go build -o /bin/checkpointz ./cmd/checkpointz

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY example_config.yaml /app/config.yaml
COPY --from=builder /bin/checkpointz /usr/local/bin/
EXPOSE 5555
ENTRYPOINT ["checkpointz"]
