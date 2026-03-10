# Stage 1: build binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server/main.go

# Stage 2: image minimal
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=builder /app/server .
RUN mkdir -p uploads && chown -R app:app /app
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["./server"]
