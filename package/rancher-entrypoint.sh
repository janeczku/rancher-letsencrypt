#!/bin/bash

# first arg is `-flag` or `--some-flag`
if [ "${1:0:1}" = '-' ]; then
	set -- /usr/bin/rancher-letsencrypt "$@"
fi

# no argument
if [ -z "$1" ]; then
	set -- /usr/bin/rancher-letsencrypt
fi

/usr/bin/update-rancher-ssl

exec "$@"
