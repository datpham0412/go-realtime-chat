# Build the backend
FROM golang:1.20-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy backend binary
COPY --from=builder /app/main /app/main

# Expose backend port
EXPOSE 8080

# Start the backend service
CMD ["./main"]
