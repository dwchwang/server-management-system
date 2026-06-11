FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY shared/ ./shared/
COPY monitor-service/ ./monitor-service/

WORKDIR /app/monitor-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/monitor-service ./cmd/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/monitor-service .

EXPOSE 8083

CMD ["./monitor-service"]
