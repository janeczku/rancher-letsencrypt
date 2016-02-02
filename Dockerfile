FROM alpine:3.3
MAINTAINER Jan Broer <jan@festplatte.eu.org>

RUN apk add --no-cache ca-certificates

ENV LETSENCRYPT_RELEASE v0.2.5

ADD https://github.com/janeczku/rancher-letsencrypt/releases/download/${LETSENCRYPT_RELEASE}/rancher-letsencrypt-linux-amd64.tar.gz /tmp/rancher-letsencrypt.tar.gz
RUN tar -zxvf /tmp/rancher-letsencrypt.tar.gz -C /usr/bin \
	&& chmod +x /usr/bin/rancher-letsencrypt

VOLUME /etc/letsencrypt

ENTRYPOINT ["/usr/bin/rancher-letsencrypt"]