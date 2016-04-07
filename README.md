![Rancher + Let's Encrypt = Awesome Sauce](https://raw.githubusercontent.com/janeczku/rancher-letsencrypt/master/hero.png)

# Let's Encrypt Certificate Manager for Rancher

[![Circle CI](https://circleci.com/gh/janeczku/rancher-letsencrypt.svg?style=shield&circle-token=cd06c9a78ae3ef7b6c1387067c36360f62d97b7a)](https://circleci.com/gh/janeczku/rancher-letsencrypt)

A [Rancher](http://rancher.com/rancher/) service that obtains free SSL/TLS certificates from the [Let's Encrypt CA](https://letsencrypt.org/), adds them to Rancher's certificate store and manages renewal and propagation of updated certificates to load balancers.

#### Requirements
* Rancher Server >= v0.63.0
* Existing account with one of the supported DNS providers:
  * `CloudFlare`
  * `DigitalOcean`
  * `AWS Route 53`
  * `DNSimple`
  * `Dyn`

### How to use

This application is distributed via the [Rancher Community Catalog](https://github.com/rancher/community-catalog).

Enable the Community Catalog under `Admin` => `Settings` in the Rancher UI.
Then find the `Let's Encrypt` template in the Catalog section of the UI and follow the instructions.

### Building the image

`make build && make image`

### Contributions

PR's welcome!
