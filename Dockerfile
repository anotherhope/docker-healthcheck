FROM alpine:latest as runner

ADD https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh /tmp/install.sh

RUN apk add --no-cache docker make alpine-sdk go zsh

WORKDIR /tmp/healthcheck
COPY / /tmp/healthcheck

RUN make install

ENTRYPOINT [ "zsh" ]