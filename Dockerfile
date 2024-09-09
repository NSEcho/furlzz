FROM ubuntu:latest

LABEL maintainer="NSEcho"

ARG TARGETOS

ARG TARGETARCH

ENV CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH \
  GOV=go1.23.1 FRIDAVERSION=16.5.1 FRIDAOS=$TARGETOS FRIDAARCH=$TARGETARCH

RUN  apt-get update \
  && apt-get install -y wget xz-utils gcc

RUN wget https://go.dev/dl/$GOV.$TARGETOS-$TARGETARCH.tar.gz \
  && rm -rf /usr/local/go \
  && tar -C /usr/local -xzf $GOV.$TARGETOS-$TARGETARCH.tar.gz

RUN mkdir /tmp/frida-core-devkit && cd /tmp/frida-core-devkit \
  && wget https://github.com/frida/frida/releases/download/$FRIDAVERSION/frida-core-devkit-$FRIDAVERSION-$FRIDAOS-$FRIDAARCH.tar.xz -O - \
  | tar --extract --xz

RUN cp /tmp/frida-core-devkit/libfrida-core.a /usr/local/lib \
  && cp /tmp/frida-core-devkit/frida-core.h /usr/local/include

RUN mkdir /furlzz

COPY . /furlzz

RUN cd /furlzz && /usr/local/go/bin/go build

ENTRYPOINT ["/furlzz/furlzz"]
