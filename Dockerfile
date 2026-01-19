FROM alpine:3.23.2

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
	&& apk add --no-cache tzdata \
	&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

RUN addgroup -g 1002 app \
	&& adduser -S -D -u 1002 -G app app

COPY --chown=app:app --chmod=500 order /home/app/product

WORKDIR /home/app

CMD [ "./product" ]