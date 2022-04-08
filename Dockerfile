FROM golang:1.18-alpine3.15 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /tyui cmd/server/*.go

FROM alpine:3.15

COPY --from=builder /tyui /

CMD ["/tyui"]