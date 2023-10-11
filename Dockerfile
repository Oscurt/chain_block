FROM golang:1.21.1-alpine as builder

WORKDIR /app

COPY . .

RUN go mod tidy \
&& go build -o main .

FROM alpine:3.18.4

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]