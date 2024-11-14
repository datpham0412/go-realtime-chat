# Stage 1: Build the frontend
FROM node:18-alpine as frontend-build
WORKDIR /app/frontend

# Install build dependencies
RUN apk add --no-cache python3 make g++

# Copy frontend files
COPY frontend/package*.json ./
RUN npm install --legacy-peer-deps

# Copy frontend source
COPY frontend/ .

# Build frontend with proper public path
ENV NODE_ENV=production
ENV PUBLIC_URL=/
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.22-alpine as backend-build
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first
COPY go.mod go.sum ./

# Update and download dependencies
RUN go mod download && \
    go mod verify

# Copy backend source
COPY . .

# Copy frontend build to static directory
COPY --from=frontend-build /app/frontend/dist ./static

# Update modules and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 3: Final image
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy built assets and binary
COPY --from=frontend-build /app/frontend/dist ./static
COPY --from=backend-build /app/main .

# Set production environment
ENV NODE_ENV=production
ENV PORT=8080

# Expose port
EXPOSE 8080

# Start application
CMD ["./main"]