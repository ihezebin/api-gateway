FROM alpine

WORKDIR /root

COPY build/${PROJECT_NAME} /root/
COPY config/config.toml /root/config/
COPY config/blog.toml /root/config/
COPY config/sso.toml /root/config/

CMD ["/root/api-gateway", "-c", "/root/config/config.toml"]
