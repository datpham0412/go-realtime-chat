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

# Build frontend
ENV NODE_ENV=production
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.22-alpine as backend-build
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy and download backend dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Build backend
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 3: Final image
FROM alpine:latest
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy built assets
COPY --from=frontend-build /app/frontend/dist ./static
COPY --from=backend-build /app/main .

# Expose port
EXPOSE 8080

# Start application
CMD ["./main"]