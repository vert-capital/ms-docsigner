FROM golang:1.25-alpine as builder

ARG PROJECT_NAME="app"
ARG VERSION="dev"
ARG COMMIT="none"
ENV CGO_ENABLED=0
ENV GOPROXY=https://proxy.golang.org

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app_bin .

FROM debian:stable

WORKDIR /app


COPY --from=builder /app/app_bin /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


EXPOSE 8080

CMD ["/app/app_bin"]