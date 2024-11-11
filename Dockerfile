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

# Install Node.js
RUN apk add --no-cache nodejs npm

# Copy frontend build
COPY --from=frontend-build /app/frontend/dist /app/static

# Copy backend binary
COPY --from=backend-build /app/main /app/main

# Expose ports
EXPOSE 8080

# Start the application
CMD ["./main"]