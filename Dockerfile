#FROM golang
FROM alpine:latest

MAINTAINER  jzy
WORKDIR /go/src/stress-testing-service
# 必须配置，windows平台制作镜像可运行，centos不添加如下代码报 /bin/sh: /go/src/stress-testing-service/stress-testing-service: not found
RUN mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY stress-testing-service /go/src/stress-testing-service/stress-testing-service
COPY config/*.json /go/src/stress-testing-service/config/

CMD /go/src/stress-testing-service/stress-testing-service ${GO_OPTS}
