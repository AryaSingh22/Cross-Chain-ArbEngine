# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod ./
RUN go mod download 2>/dev/null || true

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /arbengine ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /arbengine .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 9090

CMD ["./arbengine"]
