FROM gliderlabs/alpine:3.3
MAINTAINER Jorge Lorenzo <jlorgal@gmail.com>

RUN apk add --no-cache bash git go

RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

# Force a foreground task to prevent from stopping the container
CMD ["tail", "-f", "/dev/null"]
