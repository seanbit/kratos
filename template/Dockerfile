FROM golang:1.24.5-alpine3.21 AS builder
ARG COMMON_PAT_TOKEN
ARG APP_NAME
ARG RUN_TYPE
ARG JOB_NAME

ENV NAME=${APP_NAME} \
    ENV_RUN_TYPE=${RUN_TYPE} \
    ENV_JOB_NAME=${JOB_NAME} \
    GOPROXY=https://goproxy.org \
    GO111MODULE="on" \
    GOPRIVATE="github.com/carv-protocol"

RUN echo "APP_NAME:${APP_NAME}"
RUN echo "NAME:${NAME}"
RUN echo "RUN_TYPE:${RUN_TYPE}"
RUN echo "JOB_NAME:${JOB_NAME}"
RUN echo "ENV_RUN_TYPE:${ENV_RUN_TYPE}"
RUN echo "ENV_JOB_NAME:${ENV_JOB_NAME}"

WORKDIR /data

COPY . .

RUN apk update && \
    apk upgrade && \
    apk add --no-cache curl bash git binutils vim gdb openssh-client gcc g++ make libffi-dev openssl-dev libtool protobuf&& \
    echo 'set auto-load safe-path /' > /root/.gdbinit && \
    git config --global --add url."https://${COMMON_PAT_TOKEN}:x-oauth-basic@github.com".insteadOf "https://github.com"


RUN go mod download && make build


FROM alpine:3.18.3
WORKDIR /app
RUN wget https://public.carv.io/game/banana/start.webp
ARG APP_NAME
ARG ENV
ARG RUN_TYPE
ARG JOB_NAME

ENV PATH="/usr/local/go/bin:${PATH}" \
    GOPRIVATE="github.com/carv-protocol" \
    GO111MODULE="on" \
    ENV=${ENV} \
    NAME=${APP_NAME} \
    ENV_RUN_TYPE=${RUN_TYPE} \
    ENV_JOB_NAME=${JOB_NAME}

RUN apk update && \
    apk add --no-cache curl netcat-openbsd bind-tools

COPY --from=builder /data/bin /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["sh", "-c", "./server -conf /data/conf -run-type ${ENV_RUN_TYPE} -job-name ${ENV_JOB_NAME}"]
