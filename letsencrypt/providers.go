package letsencrypt

import (
	"fmt"
	"os"

	lego "github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/cloudflare"
	"github.com/xenolf/lego/providers/dns/digitalocean"
	"github.com/xenolf/lego/providers/dns/dnsimple"
	"github.com/xenolf/lego/providers/dns/dyn"
	"github.com/xenolf/lego/providers/dns/ovh"
	"github.com/xenolf/lego/providers/dns/route53"
	"github.com/xenolf/lego/providers/dns/vultr"
)

// ProviderOpts is used to configure the DNS provider
// used by the Let's Encrypt client for domain validation
type ProviderOpts struct {
	Provider Provider

	// CloudFlare credentials
	CloudflareEmail string
	CloudflareKey   string

	// DigitalOcean credentials
	DoAccessToken string

	// AWS Route 53 credentials
	AwsAccessKey string
	AwsSecretKey string

	// DNSimple credentials
	DNSimpleEmail string
	DNSimpleKey   string

	// Dyn credentials
	DynCustomerName string
	DynUserName     string
	DynPassword     string

	// Vultr credentials
	VultrApiKey string

	// OVH credentials
	OvhApplicationKey    string
	OvhApplicationSecret string
	OvhConsumerKey       string

	// HTTP challenge options
	HTTPWebrootPath string
}

type Provider string

const (
	CLOUDFLARE   = Provider("CloudFlare")
	DIGITALOCEAN = Provider("DigitalOcean")
	ROUTE53      = Provider("Route53")
	DNSIMPLE     = Provider("DNSimple")
	DYN          = Provider("Dyn")
	VULTR        = Provider("Vultr")
	OVH          = Provider("Ovh")
	HTTP         = Provider("HTTP")
)

type ProviderFactory struct {
	factory interface{}
	challenge lego.Challenge
}

var providerFactory = map[Provider]ProviderFactory{
	CLOUDFLARE:   ProviderFactory{makeCloudflareProvider,   lego.DNS01},
	DIGITALOCEAN: ProviderFactory{makeDigitalOceanProvider, lego.DNS01},
	ROUTE53:      ProviderFactory{makeRoute53Provider,      lego.DNS01},
	DNSIMPLE:     ProviderFactory{makeDNSimpleProvider,     lego.DNS01},
	DYN:          ProviderFactory{makeDynProvider,          lego.DNS01},
	VULTR:        ProviderFactory{makeVultrProvider,        lego.DNS01},
	OVH:          ProviderFactory{makeOvhProvider,          lego.DNS01},
	HTTP:         ProviderFactory{makeHTTPProvider,         lego.HTTP01},
}

func getProvider(opts ProviderOpts) (lego.ChallengeProvider, lego.Challenge, error) {
	if f, ok := providerFactory[opts.Provider]; ok {
		provider, err := f.factory.(func(ProviderOpts) (lego.ChallengeProvider, error))(opts)
		if err != nil {
			return nil, f.challenge, err
		}
		return provider, f.challenge, nil
	}
	irrelevant := lego.DNS01
	return nil, irrelevant, fmt.Errorf("Unsupported provider: %s", opts.Provider)
}

// returns a preconfigured CloudFlare lego.ChallengeProvider
func makeCloudflareProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.CloudflareEmail) == 0 {
		return nil, fmt.Errorf("CloudFlare email is not set")
	}
	if len(opts.CloudflareKey) == 0 {
		return nil, fmt.Errorf("CloudFlare API key is not set")
	}

	provider, err := cloudflare.NewDNSProviderCredentials(opts.CloudflareEmail, opts.CloudflareKey)
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

	provider, err := digitalocean.NewDNSProviderCredentials(opts.DoAccessToken)
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

	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY_ID", opts.AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", opts.AwsSecretKey)

	provider, err := route53.NewDNSProvider()
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured DNSimple lego.ChallengeProvider
func makeDNSimpleProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.DNSimpleEmail) == 0 {
		return nil, fmt.Errorf("DNSimple Email is not set")
	}
	if len(opts.DNSimpleKey) == 0 {
		return nil, fmt.Errorf("DNSimple API key is not set")
	}

	provider, err := dnsimple.NewDNSProviderCredentials(opts.DNSimpleEmail, opts.DNSimpleKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured Dyn lego.ChallengeProvider
func makeDynProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.DynCustomerName) == 0 {
		return nil, fmt.Errorf("Dyn customer name is not set")
	}
	if len(opts.DynUserName) == 0 {
		return nil, fmt.Errorf("Dyn user name is not set")
	}
	if len(opts.DynPassword) == 0 {
		return nil, fmt.Errorf("Dyn password is not set")
	}

	provider, err := dyn.NewDNSProviderCredentials(opts.DynCustomerName,
		opts.DynUserName, opts.DynPassword)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured Vultr lego.ChallengeProvider
func makeVultrProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.VultrApiKey) == 0 {
		return nil, fmt.Errorf("Vultr API key is not set")
	}

	provider, err := vultr.NewDNSProviderCredentials(opts.VultrApiKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured Ovh lego.ChallengeProvider
func makeOvhProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.OvhApplicationKey) == 0 {
		return nil, fmt.Errorf("OVH application key is not set")
	}
	if len(opts.OvhApplicationSecret) == 0 {
		return nil, fmt.Errorf("OVH application secret is not set")
	}
	if len(opts.OvhConsumerKey) == 0 {
		return nil, fmt.Errorf("OVH consumer key is not set")
	}

	provider, err := ovh.NewDNSProviderCredentials("ovh-eu", opts.OvhApplicationKey, opts.OvhApplicationSecret,
		opts.OvhConsumerKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured HTTP lego.ChallengeProvider
func makeHTTPProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	provider := lego.NewHTTPProviderServer("", "")

	return provider, nil
}
