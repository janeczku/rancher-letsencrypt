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
* Rancher Server >= v1.2.0
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

* If using the HTTP challenge, a proxy that routes `example.com/.well-known/acme-challenge` to `rancher-letsencrypt`.

### How to use

This application is distributed via the [Rancher Community Catalog](https://github.com/rancher/community-catalog).

Enable the Community Catalog under `Admin` => `Settings` in the Rancher UI.
Then locate the `Let's Encrypt` template in the Catalog section of the UI and follow the instructions.

### Storing certificate in shared storage volume

By default the created SSL certificate is stored in Rancher for usage in load balancers.  

If you specify an existing volume storage driver (e.g. rancher-nfs) then the account data, certificate and private key will be stored in a stack scoped volume named `lets-encrypt`, allowing you to access them from other services in the same stack. See the [Storage Service documentation](https://docs.rancher.com/rancher/v1.3/en/rancher-services/storage-service/).

#### Example

When mounting the `lets-encrypt` storage volume to `/etc/letsencrypt` in another container, then production certificates and keys are located at:
 
- `/etc/letsencrypt/production/certs/<certificate name>/fullchain.pem`
- `/etc/letsencrypt/production/certs/<certificate name>/privkey.pem`

where `<certificate name>` is the name of the certificate sanitized to consist of only the following characters: `[a-zA-Z0-9-_.]`.

### Provider specific usage

#### AWS Route 53

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
Then make sure that HTTP requests to `domain.com/.well-known/acme-challenge` are forwarded to the `rancher-letsencrypt` service, e.g. by configuring a Rancher load balancer accordingly.

![Rancher Load Balancer Let's Encrypt Targets](https://cloud.githubusercontent.com/assets/198988/22224463/0d1eb4aa-e1bf-11e6-955c-5f0d085ce8cd.png)

### Building the image

`make build && make image`

### Contributions

PR's welcome!
