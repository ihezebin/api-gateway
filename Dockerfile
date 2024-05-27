FROM alpine

WORKDIR /root

COPY build/${PROJECT_NAME} /root/
COPY config/config.toml /root/

CMD ["/root/api-gateway", "-c", "/root/config.toml"]
