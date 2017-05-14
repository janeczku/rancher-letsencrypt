FROM alpine:3.5

RUN apk add --no-cache ca-certificates openssl bash

ENV LETSENCRYPT_RELEASE v0.5.0
ENV SSL_SCRIPT_COMMIT 08278ace626ada71384fc949bd637f4c15b03b53

RUN wget -O /usr/bin/update-rancher-ssl https://raw.githubusercontent.com/rancher/rancher/${SSL_SCRIPT_COMMIT}/server/bin/update-rancher-ssl && \
    chmod +x /usr/bin/update-rancher-ssl

COPY rancher-entrypoint.sh /usr/bin/

ADD https://github.com/janeczku/rancher-letsencrypt/releases/download/${LETSENCRYPT_RELEASE}/rancher-letsencrypt-linux-amd64.tar.gz /tmp/rancher-letsencrypt.tar.gz

RUN tar -zxvf /tmp/rancher-letsencrypt.tar.gz -C /usr/bin \
	&& chmod +x /usr/bin/rancher-letsencrypt

EXPOSE 80
ENTRYPOINT ["/usr/bin/rancher-entrypoint.sh"]
