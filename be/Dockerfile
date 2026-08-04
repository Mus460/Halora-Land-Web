# syntax=docker/dockerfile:1
FROM golang:1.26-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/halora-be ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 1001 halora
WORKDIR /app

COPY --from=builder /out/halora-be /usr/local/bin/halora-be
COPY data ./data

USER halora
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["halora-be"]
CMD ["serve"]
