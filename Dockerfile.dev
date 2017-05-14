FROM alpine:3.5

RUN apk add --no-cache ca-certificates

ADD build/rancher-letsencrypt-linux-amd64 /usr/bin/rancher-letsencrypt

RUN chmod +x /usr/bin/rancher-letsencrypt

ENTRYPOINT ["/usr/bin/rancher-letsencrypt", "-debug", "-test-mode"]
EXPOSE 80