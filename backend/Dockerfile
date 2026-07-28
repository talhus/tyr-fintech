FROM golang:1.26-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]