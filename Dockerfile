# Stage 1: Build the frontend
FROM node:16-alpine as frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.20-alpine as backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Stage 3: Final image
FROM alpine:latest
WORKDIR /app

# Install Redis and Node.js
RUN apk add --no-cache redis nodejs npm

# Copy frontend build
COPY --from=frontend-build /app/frontend/dist /app/frontend/dist

# Install serve globally
RUN npm install -g serve

# Copy backend binary
COPY --from=backend-build /app/main /app/main

# Create start script (fixed version)
RUN printf '#!/bin/sh\nredis-server --daemonize yes\nserve -s frontend/dist -l 3000 &\n./main' > /app/start.sh && \
    chmod +x /app/start.sh

# Expose ports
EXPOSE 3000 8080 6379

# Start all services
CMD ["/app/start.sh"]
