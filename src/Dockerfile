FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

ARG GRAFANA_URL
ARG GRAFANA_TOKEN

RUN CGO_ENABLED=0 GOOS=linux go build -o calc-pi main.go


FROM alpine:latest

COPY --from=builder /app/calc-pi /

CMD ["./calc-pi"]