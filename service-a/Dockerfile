FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go mod tidy
RUN go build -o service-a main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/service-a .

EXPOSE 8080

CMD ["./service-a"]
