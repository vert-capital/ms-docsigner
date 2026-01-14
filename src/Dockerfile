FROM golang:1.25-alpine AS builder

ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY . .

RUN apk update && apk add --no-cache gcc libc-dev librdkafka-dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app_bin .

FROM debian:bullseye

RUN apt-get update -y && \
    apt-get install -y wget gnupg2 && \
    echo "deb http://apt.postgresql.org/pub/repos/apt bullseye-pgdg main" > /etc/apt/sources.list.d/pgdg.list && \
    wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - && \
    apt-get update -y && \
    apt-get install -y postgresql-client-14

WORKDIR /app

COPY --from=builder /app/app_bin /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/app/app_bin"]