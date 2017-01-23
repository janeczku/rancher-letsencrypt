![Rancher + Let's Encrypt = Awesome Sauce](https://raw.githubusercontent.com/janeczku/rancher-letsencrypt/master/hero.png)

# Let's Encrypt Certificate Manager for Rancher

[![Latest Version](https://img.shields.io/github/release/janeczku/rancher-letsencrypt.svg?maxAge=2592000)][release]
[![Circle CI](https://circleci.com/gh/janeczku/rancher-letsencrypt.svg?style=shield&circle-token=cd06c9a78ae3ef7b6c1387067c36360f62d97b7a)][circleci]
[![Docker Pulls](https://img.shields.io/docker/pulls/janeczku/rancher-letsencrypt.svg?maxAge=2592000)][hub]
[![License](https://img.shields.io/github/license/janeczku/rancher-letsencrypt.svg?maxAge=2592000)]()

[release]: https://github.com/janeczku/rancher-letsencrypt/releases
[circleci]: https://circleci.com/gh/janeczku/rancher-letsencrypt
[hub]: https://hub.docker.com/r/janeczku/rancher-letsencrypt/


A [Rancher](http://rancher.com/rancher/) service that obtains free SSL/TLS certificates from the [Let's Encrypt CA](https://letsencrypt.org/), adds them to Rancher's certificate store and manages renewal and propagation of updated certificates to load balancers.

#### Requirements
* Rancher Server >= v0.63.0
* If using a DNS-based challenge, existing account with one of the supported DNS providers:
  * `AWS Route 53`
  * `CloudFlare`
  * `DigitalOcean`
  * `DNSimple`
  * `Dyn`
  * `Vultr`
  * `Ovh`
* If using the HTTP challenge, a proxy that routes `example.com/.well-known/acme-challenge` to `rancher-letsencrypt`.

### How to use

This application is distributed via the [Rancher Community Catalog](https://github.com/rancher/community-catalog).

Enable the Community Catalog under `Admin` => `Settings` in the Rancher UI.
Then locate the `Let's Encrypt` template in the Catalog section of the UI and follow the instructions.

#### Accessing certificates and private keys from other services
The created SSL certificate is stored in Rancher for usage in load balancers.    
If you want to use it from other services (e.g. a Nginx container) you can opt to save the certificate and private key to a host path,
named volume or Convoy storage volume. You can then mount the volume or host path to other containers and access the files as follows:    
`<path_on_host or volume name>/<certificate name>/fullchain.pem`    
`<path_on_host or volume name>/<certificate name>/privkey.pem`    
where `<certificate name>` is the name you specified in the UI forced to this set of characters: `[a-zA-Z0-9-_.]`.



### Provider specific usage

#### AWS Route 53

The following IAM policy describes the minimum permissions required to run `rancher-letsencrypt` using AWS Route 53 for domain authorization.    
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

You need to create your credential on the following URL: https://eu.api.ovh.com/createToken/
Then submit the form as following:
- `Account ID`: Your OVH account ID
- `Password`: Your password
- `Script name`: letsencrypt
- `Script description`: Letsencrypt for Rancher
- `Validity`: Unlimited
- `Rights`:
  - GET /domain/zone/*
  - POST /domain/zone/*
  - DELETE /domain/zone/*

Then get your key and store them.

To finish, when you start this container add the following environment variable:
- `PROVIDER`: Ovh
- `OVH_APPLICATION_KEY`: your key generated in previous step
- `OVH_APPLICATION_SECRET`: your secret generated in previous step
- `OVH_CONSUMER_KEY`: your consumer key generated in previous step

#### HTTP

If you prefer not to use a DNS-based challenge
or your provider is not supported, you can use the HTTP challenge.

Simply set the following option:
- `PROVIDER`: HTTP

With this you'll have to make sure that HTTP
requests to `example.com/.well-known/acme-challenge` get redirected
to your `rancher-letsencrypt` instance. You can use a reverse proxy, like
the Rancher Load Balancer for that:

![Rancher Load Balancer LetsEncrypt Targets](https://cloud.githubusercontent.com/assets/198988/22224463/0d1eb4aa-e1bf-11e6-955c-5f0d085ce8cd.png)

### Building the image

`make build && make image`

### Contributions

PR's welcome!
