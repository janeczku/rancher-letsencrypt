![Rancher + Let's Encrypt = Awesome Sauce](https://raw.githubusercontent.com/janeczku/rancher-letsencrypt/master/hero.png)

# Let's Encrypt Certificate Manager for Rancher

[![Circle CI](https://circleci.com/gh/janeczku/rancher-letsencrypt.svg?style=shield&circle-token=cd06c9a78ae3ef7b6c1387067c36360f62d97b7a)](https://circleci.com/gh/janeczku/rancher-letsencrypt)

A [Rancher](http://rancher.com/rancher/) service that obtains free SSL/TLS certificates from the [Let's Encrypt CA](https://letsencrypt.org/), adds them to Rancher's certificate store and manages renewal and propagation of updated certificates to load balancers.

#### Requirements
* Rancher Server >= v0.63.0
* Existing account with one of the supported DNS providers:
  * `AWS Route 53`
  * `CloudFlare`
  * `DigitalOcean`
  * `DNSimple`
  * `Dyn`
  * `Namecheap`

### How to use

This application is distributed via the [Rancher Community Catalog](https://github.com/rancher/community-catalog).

Enable the Community Catalog under `Admin` => `Settings` in the Rancher UI.
Then find the `Let's Encrypt` template in the Catalog section of the UI and follow the instructions.

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

#### Namecheap

Namecheap requires all IP addresses from which you call it's API to be whitelisted. Make sure to grant API access to the host running `rancher-letsencrypt` by navigating to "Manage Profile" => "API Access" in your Namecheap account.   
Be aware that Namecheap can be slow to propagate DNS changes (up to 60 minutes). This may slow down the process of creating certificates significantly.

### Building the image

`make build && make image`

### Contributions

PR's welcome!
