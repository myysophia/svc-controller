FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rabbitmq-controller .

FROM alpine:3.14
WORKDIR /root/
COPY --from=builder /app/rabbitmq-controller .
CMD ["./rabbitmq-controller"]