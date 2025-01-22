FROM docker.io/golang:latest AS builder

WORKDIR /app

COPY . .

RUN make build

FROM scratch

COPY --from=builder /app/coin-cache-service .

ENTRYPOINT ["/coin-cache-service"]