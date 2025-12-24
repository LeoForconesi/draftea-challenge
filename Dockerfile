FROM golang:1.24.11-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o api ./cmd/api
RUN go build -o relay ./cmd/relay
RUN go build -o consumer ./cmd/consumer

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/relay .
COPY --from=builder /app/consumer .
COPY --from=builder /app/config ./config

CMD ["./api"]
