FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o boid main.go


FROM alpine:latest

WORKDIR /

COPY --from=builder /app/boid /

RUN apk add --no-cache ca-certificates

CMD ["./boid"]