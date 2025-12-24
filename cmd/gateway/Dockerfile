FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway ./cmd/gateway

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/gateway /app/gateway

EXPOSE 8080

CMD ["./gateway"]

