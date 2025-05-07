FROM golang:1.24.3 AS builder

WORKDIR /chalk

COPY ./ ./
RUN mkdir -p build/bin && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./build/bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /chalk

COPY --from=builder /chalk/build/bin/app /usr/local/bin/chalk
COPY --from=builder /chalk/migrations /etc/chalk/migrations

CMD ["chalk"]