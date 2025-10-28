FROM alpine:3.22.2

ARG CONSUL_HOST=127.0.0.1
ARG CONSUL_PORT=8500
ARG CONSUL_PREFIX=product

ENV CONSUL_HOST=$CONSUL_HOST \
	CONSUL_PORT=$CONSUL_PORT \
	CONSUL_PREFIX=$CONSUL_PREFIX

COPY product /var/product

RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && chmod +x /var/product

WORKDIR /var

CMD [ "./product" ]