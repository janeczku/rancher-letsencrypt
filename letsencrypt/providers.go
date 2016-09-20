package letsencrypt

import (
	"fmt"
	"os"

	lego "github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/cloudflare"
	"github.com/xenolf/lego/providers/dns/digitalocean"
	"github.com/xenolf/lego/providers/dns/dnsimple"
	"github.com/xenolf/lego/providers/dns/dyn"
	"github.com/xenolf/lego/providers/dns/gandi"
	"github.com/xenolf/lego/providers/dns/ovh"
	"github.com/xenolf/lego/providers/dns/route53"
	"github.com/xenolf/lego/providers/dns/vultr"
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

	// Gandi credentials
	GandiApiKey string
}

type DnsProvider string

const (
	CLOUDFLARE   = DnsProvider("CloudFlare")
	DIGITALOCEAN = DnsProvider("DigitalOcean")
	ROUTE53      = DnsProvider("Route53")
	DNSIMPLE     = DnsProvider("DNSimple")
	DYN          = DnsProvider("Dyn")
	VULTR        = DnsProvider("Vultr")
	OVH          = DnsProvider("Ovh")
	GANDI        = DnsProvider("Gandi")
)

var dnsProviderFactory = map[DnsProvider]interface{}{
	CLOUDFLARE:   makeCloudflareProvider,
	DIGITALOCEAN: makeDigitalOceanProvider,
	ROUTE53:      makeRoute53Provider,
	DNSIMPLE:     makeDNSimpleProvider,
	DYN:          makeDynProvider,
	VULTR:        makeVultrProvider,
	OVH:          makeOvhProvider,
	GANDI:        makeGandiProvider,
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

// returns a preconfigured Gandi lego.ChallengeProvider
func makeGandiProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.GandiApiKey) == 0 {
		return nil, fmt.Errorf("Gandi API key is not set")
	}

	provider, err := gandi.NewDNSProviderCredentials(opts.GandiApiKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
