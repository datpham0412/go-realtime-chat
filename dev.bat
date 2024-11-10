@echo off

IF "%1"=="redis" (
    docker run -d -p 6379:6379 --name redis-graphql redis:5.0.7
    GOTO :EOF
)

IF "%1"=="redis-stop" (
    docker stop redis-graphql
    docker rm redis-graphql
    GOTO :EOF
)

IF "%1"=="dev-backend" (
    go run main.go
    GOTO :EOF
)

IF "%1"=="dev-frontend" (
    cd frontend && npm start
    GOTO :EOF
)

echo Available commands:
echo dev redis        - Start Redis in Docker
echo dev redis-stop   - Stop Redis container
echo dev dev-backend  - Run the backend
echo dev dev-frontend - Run the frontend