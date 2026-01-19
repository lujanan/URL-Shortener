FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o url-shortener ./cmd/server

FROM alpine:3.19

WORKDIR /app

RUN adduser -D appuser
USER appuser

COPY --from=builder /app/url-shortener /app/url-shortener

EXPOSE 8080

ENV PORT=8080
ENV DB_DSN="shortuser:shortpass@tcp(mysql:3306)/shorturl?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

ENTRYPOINT ["/app/url-shortener"]

