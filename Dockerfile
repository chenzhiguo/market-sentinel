FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o sentinel ./cmd/sentinel

# Runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=builder /app/sentinel .

# Copy configs
COPY configs ./configs

# Create data directory
RUN mkdir -p /app/data/reports

# Environment
ENV TZ=Asia/Shanghai
ENV SENTINEL_SERVER_HOST=0.0.0.0
ENV SENTINEL_SERVER_PORT=8080

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./sentinel"]
CMD ["serve"]
