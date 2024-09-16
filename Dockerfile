FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o tender-service /app/cmd/main

FROM ubuntu:22.04
WORKDIR /app
COPY --from=builder /app/tender-service .
EXPOSE 8080
CMD ["./tender-service"]