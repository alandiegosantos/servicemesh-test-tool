FROM golang:alpine as builder

LABEL maintainer="alandiegosantos@gmail.com"

RUN mkdir /build 
ADD . /build/
WORKDIR /build 
RUN apk update && apk add make && make clean && make

FROM alpine
RUN adduser -S -D -H appuser
USER appuser
COPY --from=builder /build/webserver /usr/local/bin/webserver
EXPOSE 8080
CMD ["/usr/local/bin/webserver","--conf","/etc/dependencies.yaml"]