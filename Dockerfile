FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o datapipe .

FROM alpine:3.20.3
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/datapipe .
EXPOSE 8080
CMD ["./datapipe"]
