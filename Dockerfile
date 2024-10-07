# Stage 1: Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

# Build the Go app
RUN go mod init go-docker-app && \
    go mod tidy && \
    go build -o /go-docker-app

# Stage 2: Run stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the build stage
COPY --from=builder /go-docker-app .

# Run the Go app
CMD ["./go-docker-app"]
