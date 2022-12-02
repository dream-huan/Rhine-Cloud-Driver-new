FROM ubuntu:focal

ENV PROC_NAME rhine-cloud-driver

EXPOSE 8888

RUN chmod +x ./start.sh
CMD ./start.sh

# # 配置时区
# RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
#     && echo "Asia/Shanghai" > /etc/timezone

# WORKDIR /rhine-cloud-driver

# VOLUME [ "/rhine-cloud-driver/uploads", "/data"]

# ENTRYPOINT [ "./rhine-cloud-driver" ]