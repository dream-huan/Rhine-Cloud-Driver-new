# 取ubuntu最新镜像
FROM scratch
ADD ubuntu-focal-oci-amd64-root.tar.gz /
CMD ["bash"]

ENV PROC_NAME rhine-cloud-driver

# 开放端口

EXPOSE 8888

