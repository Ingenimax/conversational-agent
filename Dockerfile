# Stage 1: Build the application
FROM golang:1.23.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/main.go

# Stage 2: Run the application
FROM debian:bullseye-slim

WORKDIR /app

RUN apt update \
    && apt install -y ca-certificates \
    && update-ca-certificates

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]