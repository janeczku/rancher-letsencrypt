FROM alpine:3.3
MAINTAINER <jan@rancher.com>

RUN apk add --no-cache ca-certificates

ENV LETSENCRYPT_RELEASE v0.3.0

ADD https://github.com/janeczku/rancher-letsencrypt/releases/download/${LETSENCRYPT_RELEASE}/rancher-letsencrypt-linux-amd64.tar.gz /tmp/rancher-letsencrypt.tar.gz

RUN echo -n "d42f6ea5f8b8ad20bff30a9689c5a71203e4f2629e05cef34c0cba67965f5012  /tmp/rancher-letsencrypt.tar.gz" | \
    sha256sum -c \
    && tar -zxvf /tmp/rancher-letsencrypt.tar.gz -C /usr/bin \
	&& chmod +x /usr/bin/rancher-letsencrypt

ENTRYPOINT ["/usr/bin/rancher-letsencrypt"]