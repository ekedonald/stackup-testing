# Build stage
FROM golang:1.23-alpine3.19 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o tracker .

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/tracker .
CMD ["./tracker"]