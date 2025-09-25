FROM golang:1.24-alpine AS builder

ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"
ENV CGO_ENABLED=1
ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY . .

RUN apk update && apk add --no-cache gcc libc-dev librdkafka-dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags musl -o /app/app_bin .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app_bin /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/app/app_bin"]