# Stage 1: Builder
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dns_udp \
    -trimpath \
    -ldflags="-s -w -X 'main.version=$(git describe --tags --always 2>/dev/null || echo $(git rev-parse --short HEAD))' -X 'main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')'" \
    ./main.go

# Stage 2: Final Image
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/dns_udp /app/
RUN apk add --no-cache libc6-compat
# COPY assets /app/assets  # Descomenta si usas assets

CMD ["/app/dns_udp"]