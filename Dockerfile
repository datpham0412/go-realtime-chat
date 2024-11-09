# Step 1: Build Stage
FROM golang:1.13-alpine3.11 AS build
RUN apk --no-cache add clang gcc g++ make git ca-certificates git

WORKDIR /go/src/github.com/datpham0412/go-realtime-chat
COPY go.mod go.sum ./                   
RUN go mod download                   

COPY . ./                              
RUN go build -o /go/bin/app .           

# Step 2: Runtime Stage
FROM alpine:3.11
WORKDIR /usr/bin
COPY --from=build /go/bin/app .         
ENV REDIS_URL=localhost:6379
EXPOSE 8080                             
CMD ["./app"]                           
