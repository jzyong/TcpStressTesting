FROM golang:1.19.2-alpine AS builder
WORKDIR /usr/src/app/

ENV GOPROXY https://goproxy.cn

COPY ./ ./

# Build executable
RUN go build -o /go/bin/tcp-stress-testing/ ./
COPY ./config/ /go/bin/tcp-stress-testing/config/


FROM alpine:latest

MAINTAINER  jzy
WORKDIR /go/src/app
# 必须配置，windows平台制作镜像可运行，centos不添加如下代码报 /bin/sh: /go/src/stress-testing-service/stress-testing-service: not found
RUN mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY --from=builder /go/bin/tcp-stress-testing/ ./

CMD ./TcpStressTesting ${OPTS}
