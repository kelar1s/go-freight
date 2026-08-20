# Этап сборки
FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/gofreight-app ./cmd/inventory/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/gofreight-app ./app
COPY --from=builder /src/config/ ./config/

EXPOSE 8080

ENTRYPOINT ["./app"]
