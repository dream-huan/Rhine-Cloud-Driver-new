FROM golang:1.19.3-alpine

RUN apk update && apk add --no-cache wget curl git

WORKDIR /rhine-cloud-driver-builder

RUN git clone --recurse-submodules https://github.com/dream-huan/Rhine-Cloud-Driver-new.git

WORKDIR /rhine-cloud-driver-builder/Rhine-Cloud-Driver-new

RUN go build

FROM alpine:latest

WORKDIR /rhine-cloud-driver

RUN apk add --no-cache bash tzdata ca-certificates && \
    update-ca-certificates

# 配置时区
RUN apk --no-cache add tzdata ca-certificates && \
    update-ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone
ENV TZ Asia/Shanghai

ENV PROC_NAME rhine-cloud-driver

EXPOSE 8888

COPY --from=0 /rhine-cloud-driver-builder/Rhine-Cloud-Driver-new/Rhine-Cloud-Driver ./

COPY --from=0 /rhine-cloud-driver-builder/Rhine-Cloud-Driver-new/wait-for-it.sh ./

VOLUME [ "/rhine-cloud-driver/uploads",  "/rhine-cloud-driver/avatar", "/rhine-cloud-driver/logs", "/rhine-cloud-driver/uploads/thumbnail"]

RUN chmod +x ./wait-for-it.sh

ENTRYPOINT [ "./wait-for-it.sh","0.0.0.0:3306","--","./Rhine-Cloud-Driver" ]