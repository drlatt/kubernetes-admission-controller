FROM golang as builder

# RUN mkdir /build
WORKDIR /build
ADD . /build/
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook_server

FROM alpine:3.15.4
WORKDIR /app
COPY --from=builder /build/webhook_server  /app
COPY --from=builder /build/ssl-certs /app/ssl-certs

CMD ["./webhook_server"]
