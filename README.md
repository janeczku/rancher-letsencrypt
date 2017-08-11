![Rancher + Let's Encrypt = Awesome Sauce](https://raw.githubusercontent.com/janeczku/rancher-letsencrypt/master/hero.png)

# Let's Encrypt Certificate Manager for Rancher

[![Latest Version](https://img.shields.io/github/release/janeczku/rancher-letsencrypt.svg?maxAge=8600)][release]
[![Circle CI](https://circleci.com/gh/janeczku/rancher-letsencrypt.svg?style=shield&circle-token=cd06c9a78ae3ef7b6c1387067c36360f62d97b7a)][circleci]
[![Docker Pulls](https://img.shields.io/docker/pulls/janeczku/rancher-letsencrypt.svg?maxAge=8600)][hub]
[![License](https://img.shields.io/github/license/janeczku/rancher-letsencrypt.svg?maxAge=8600)]()

[release]: https://github.com/janeczku/rancher-letsencrypt/releases
[circleci]: https://circleci.com/gh/janeczku/rancher-letsencrypt
[hub]: https://hub.docker.com/r/janeczku/rancher-letsencrypt/

A [Rancher](http://rancher.com/rancher/) service that obtains free SSL/TLS certificates from the [Let's Encrypt CA](https://letsencrypt.org/), adds them to Rancher's certificate store and manages renewal and propagation of updated certificates to load balancers.

#### Requirements
* Rancher Server >= v1.5.0
* If using a DNS-based challenge, existing account with one of the supported DNS providers:
  * `Aurora DNS`
  * `AWS Route 53`
  * `Azure DNS`
  * `CloudFlare`
  * `DigitalOcean`
  * `DNSimple`
  * `Dyn`
  * `Gandi`
  * `NS1`
  * `Ovh`
  * `PowerDNS`
  * `Vultr`

* If using the HTTP challenge, a reverse proxy that routes `example.com/.well-known/acme-challenge` to `rancher-letsencrypt`. 

### How to use

This application is distributed via the [Rancher Community Catalog](https://github.com/rancher/community-catalog).

Enable the Community Catalog under `Admin` => `Settings` in the Rancher UI.
Then locate the `Let's Encrypt` template in the Catalog section of the UI and follow the instructions.

### Storing certificate in shared storage volume

By default the created SSL certificate is stored in Rancher's certificate store for usage in Rancher load balancers.

You can specify a volume name to store account data, certificate and private key in a (host scoped) named Docker volume.
To share the certificates with other services you may specify a persistent storage driver (e.g. rancher-nfs).

See the README in the Rancher catalog for more information.

### Provider specific usage

#### AWS Route 53

Note: If you have both a private and public zone in Route53 for the domain, you need to run the service configured with public DNS resolvers (this is now the default).

The following IAM policy describes the minimum permissions required when using AWS Route 53 for domain authorization.    
Replace `<HOSTED_ZONE_ID>` with the ID of the hosted zone that encloses the domain(s) for which you are going to obtain certificates. You may use a wildcard (*) in place of the ID to make this policy work with all of the hosted zones associated with an AWS account.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "route53:GetChange",
                "route53:ListHostedZonesByName"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": [
                "arn:aws:route53:::hostedzone/<HOSTED_ZONE_ID>"
            ]
        }
    ]
}
```

#### OVH

First create your credentials on https://eu.api.ovh.com/createToken/ by filling out the form like this:

- `Account ID`: Your OVH account ID
- `Password`: Your password
- `Script name`: letsencrypt
- `Script description`: Letsencrypt for Rancher
- `Validity`: Unlimited
- `Rights`:
  - GET /domain/zone/*
  - POST /domain/zone/*
  - DELETE /domain/zone/*

Then deploy this service using the generated key, application secret and consumer key.

#### HTTP

If you prefer not to use a DNS-based challenge or your provider is not supported, you can use the HTTP challenge.
Simply choose `HTTP` from the list of providers.
Then make sure that HTTP requests to `domain.com/.well-known/acme-challenge` are forwarded to port 80 of the `rancher-letsencrypt` service, e.g. by configuring a Rancher load balancer accordingly. If you are using another reverse proxy (e.g. Nginx) you need to make sure it passed the original `host` header through to the backend.

![Rancher Load Balancer Let's Encrypt Targets](https://cloud.githubusercontent.com/assets/198988/22224463/0d1eb4aa-e1bf-11e6-955c-5f0d085ce8cd.png)

### Building the image

`make build && make image`

### Contributions

PR's welcome!
