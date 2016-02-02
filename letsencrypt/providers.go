package letsencrypt

import (
	"fmt"

	lego "github.com/xenolf/lego/acme"
)

// ProviderOpts is used to configure the DNS provider
// used by the Let's Encrypt client for domain validation
type ProviderOpts struct {
	Provider DnsProvider

	// CloudFlare credentials
	CloudflareEmail string
	CloudflareKey   string

	// DigitalOcean credentials
	DoAccessToken string

	// AWS Route 53 credentials
	AwsAccessKey  string
	AwsSecretKey  string
	AwsRegionName string
}

type DnsProvider string

const (
	CLOUDFLARE   = DnsProvider("CloudFlare")
	DIGITALOCEAN = DnsProvider("DigitalOcean")
	ROUTE53      = DnsProvider("Route53")
)

var dnsProviderFactory = map[DnsProvider]interface{}{
	CLOUDFLARE:   makeCloudflareProvider,
	DIGITALOCEAN: makeDigitalOceanProvider,
	ROUTE53:      makeRoute53Provider,
}

func getProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if f, ok := dnsProviderFactory[opts.Provider]; ok {
		provider, err := f.(func(ProviderOpts) (lego.ChallengeProvider, error))(opts)
		if err != nil {
			return nil, err
		}
		return provider, nil
	}
	return nil, fmt.Errorf("Unsupported DNS provider: %s", opts.Provider)
}

// returns a preconfigured CloudFlare lego.ChallengeProvider
func makeCloudflareProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.CloudflareEmail) == 0 {
		return nil, fmt.Errorf("CloudFlare email is not set")
	}
	if len(opts.CloudflareKey) == 0 {
		return nil, fmt.Errorf("CloudFlare API key is not set")
	}

	provider, err := lego.NewDNSProviderCloudFlare(opts.CloudflareEmail, opts.CloudflareKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured DigitalOcean lego.ChallengeProvider
func makeDigitalOceanProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.DoAccessToken) == 0 {
		return nil, fmt.Errorf("DigitalOcean API access token is not set")
	}

	provider, err := lego.NewDNSProviderDigitalOcean(opts.DoAccessToken)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured Route53 lego.ChallengeProvider
func makeRoute53Provider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.AwsAccessKey) == 0 {
		return nil, fmt.Errorf("AWS access key is not set")
	}
	if len(opts.AwsSecretKey) == 0 {
		return nil, fmt.Errorf("AWS secret key is not set")
	}
	if len(opts.AwsRegionName) == 0 {
		return nil, fmt.Errorf("AWS region name is not set")
	}

	provider, err := lego.NewDNSProviderRoute53(opts.AwsAccessKey, opts.AwsSecretKey, opts.AwsRegionName)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
