FROM golang:1.19.3-alpine

RUN apk update && apk add --no cache wget curl git

WORKDIR /rhine-cloud-driver-builder/backend

RUN git clone --recurse-submodules https://github.com/dream-huan/Rhine-Cloud-Driver-new.git

RUN go build

FROM alpine:latest

WORKDIR /rhine-cloud-driver

# 配置时区
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

ENV PROC_NAME rhine-cloud-driver

EXPOSE 8888

COPY --from=0 /rhine-cloud-driver-builder/backend/rhine-cloud-driver ./

VOLUME [ "/rhine-cloud-driver/uploads",  "/rhine-cloud-driver/avatar"]

ENTRYPOINT [ "./rhine-cloud-driver" ]