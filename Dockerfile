FROM gliderlabs/alpine:3.3
MAINTAINER Jorge Lorenzo <jlorgal@gmail.com>

WORKDIR /opt/gocilla

RUN mkdir -p /opt/gocilla /etc/gocilla
COPY config.json /etc/gocilla/
COPY gocilla /opt/gocilla

ENV CONFIG_PATH /etc/gocilla/gocilla.json
ENTRYPOINT ["gocilla"]

EXPOSE 3000

ARG VERSION
LABEL version="$VERSION"
