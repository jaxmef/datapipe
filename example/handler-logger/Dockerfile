FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN GO111MODULE=off go build -o handler .
EXPOSE 8082
CMD ["./handler"]
