FROM golang:1.24-alpine AS builder
ARG COMMON_PAT_TOKEN
ARG APP_NAME
ARG RUN_TYPE

ENV GOPROXY=https://goproxy.cn,direct \
    GO111MODULE=on \
    GOPRIVATE=github.com/carv-protocol

WORKDIR /data

RUN apk add --no-cache git gcc g++ make && \
    git config --global --add url."https://${COMMON_PAT_TOKEN}:x-oauth-basic@github.com".insteadOf "https://github.com"

# Leverage Docker layer cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build


FROM alpine:3.18.3
WORKDIR /app

ARG RUN_TYPE
ENV ENV_RUN_TYPE=${RUN_TYPE} \
    SECRET_FILE=""

RUN apk update && \
    apk add --no-cache curl netcat-openbsd bind-tools

COPY --from=builder /data/bin /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["sh", "-c", "./${ENV_RUN_TYPE} -conf /data/conf ${SECRET_FILE:+-secret-file $SECRET_FILE}"]
