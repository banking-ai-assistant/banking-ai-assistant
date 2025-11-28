FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/app

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates bash

COPY --from=builder /server /server

COPY migrations ./migrations

EXPOSE 8081

ENTRYPOINT ["/server"]
