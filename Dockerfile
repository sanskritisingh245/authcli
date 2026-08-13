FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/authcli ./cmd/authcli

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/authcli ./authcli
COPY migrations ./migrations
ENTRYPOINT ["./authcli"]
