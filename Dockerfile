FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /gofreight-app ./cmd/gofreight/main.go


FROM alpine:latest

WORKDIR /

COPY --from=builder /gofreight-app /gofreight-app
COPY config/ /config/

EXPOSE 8080

ENTRYPOINT ["/gofreight-app"]
