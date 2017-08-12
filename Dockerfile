FROM golang:alpine AS build-env
RUN apk add --update make git 
WORKDIR /src
ADD . .
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/annerajb/rancher-letsencrypt/letsencrypt	
RUN make build

# final stage
FROM alpine:3.5
COPY --from=build-env /src/build/rancher-letsencrypt-linux-amd64 /usr/bin/rancher-letsencrypt
COPY package/rancher-entrypoint.sh /usr/bin/
RUN apk add --no-cache ca-certificates openssl bash &&\
    wget -O /usr/bin/update-rancher-ssl https://raw.githubusercontent.com/rancher/rancher/08278ace626ada71384fc949bd637f4c15b03b53/server/bin/update-rancher-ssl && \
    chmod +x /usr/bin/update-rancher-ssl &&\
    chmod +x /usr/bin/rancher-letsencrypt

EXPOSE 80
ENTRYPOINT ["/usr/bin/rancher-entrypoint.sh"]
