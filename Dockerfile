# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /google-drive-mcp-server ./cmd/server

# Final stage
FROM alpine:latest
WORKDIR /
COPY --from=builder /google-drive-mcp-server .

EXPOSE 8080
ENTRYPOINT ["/google-drive-mcp-server"]
