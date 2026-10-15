FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/programacion-web .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/programacion-web /app/programacion-web
COPY static /app/static

EXPOSE 8080

CMD ["/app/programacion-web"]

