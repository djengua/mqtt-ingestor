FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/mqtt-ingestor ./cmd/ingestor

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /bin/mqtt-ingestor /usr/local/bin/mqtt-ingestor

EXPOSE 8080

USER app

ENTRYPOINT ["/usr/local/bin/mqtt-ingestor"]
