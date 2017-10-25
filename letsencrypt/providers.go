package letsencrypt

import (
	"fmt"
	"os"

	lego "github.com/xenolf/lego/acme"
	"github.com/xenolf/lego/providers/dns/auroradns"
	"github.com/xenolf/lego/providers/dns/azure"
	"github.com/xenolf/lego/providers/dns/cloudflare"
	"github.com/xenolf/lego/providers/dns/digitalocean"
	"github.com/xenolf/lego/providers/dns/dnsimple"
	"github.com/xenolf/lego/providers/dns/dyn"
	"github.com/xenolf/lego/providers/dns/gandi"
	"github.com/xenolf/lego/providers/dns/ns1"
	"github.com/xenolf/lego/providers/dns/ovh"
	"github.com/xenolf/lego/providers/dns/route53"
	"github.com/xenolf/lego/providers/dns/vultr"
)

// ProviderOpts is used to configure the DNS provider
// used by the Let's Encrypt client for domain validation
type ProviderOpts struct {
	Provider Provider

	// Aurora credentials
	AuroraUserId   string
	AuroraKey      string
	AuroraEndpoint string

	// AWS Route 53 credentials
	AwsAccessKey string
	AwsSecretKey string

	// Azure credentials
	AzureClientId       string
	AzureClientSecret   string
	AzureSubscriptionId string
	AzureTenantId       string
	AzureResourceGroup  string

	// CloudFlare credentials
	CloudflareEmail string
	CloudflareKey   string

	// DigitalOcean credentials
	DoAccessToken string

	// DNSimple credentials
	DNSimpleEmail string
	DNSimpleKey   string

	// Dyn credentials
	DynCustomerName string
	DynUserName     string
	DynPassword     string

	// Gandi credentials
	GandiApiKey string

	// NS1 credentials
	NS1ApiKey string

	// OVH credentials
	OvhApplicationKey    string
	OvhApplicationSecret string
	OvhConsumerKey       string

	// Vultr credentials
	VultrApiKey string
}

type Provider string

const (
	AURORA       = Provider("Aurora")
	AZURE        = Provider("Azure")
	CLOUDFLARE   = Provider("CloudFlare")
	DIGITALOCEAN = Provider("DigitalOcean")
	DNSIMPLE     = Provider("DNSimple")
	DYN          = Provider("Dyn")
	GANDI        = Provider("Gandi")
	NS1          = Provider("NS1")
	OVH          = Provider("Ovh")
	ROUTE53      = Provider("Route53")
	VULTR        = Provider("Vultr")
	HTTP         = Provider("HTTP")
)

type ProviderFactory struct {
	factory   interface{}
	challenge lego.Challenge
}

var providerFactory = map[Provider]ProviderFactory{
	AURORA:       ProviderFactory{makeAuroraProvider, lego.DNS01},
	AZURE:        ProviderFactory{makeAzureProvider, lego.DNS01},
	CLOUDFLARE:   ProviderFactory{makeCloudflareProvider, lego.DNS01},
	DIGITALOCEAN: ProviderFactory{makeDigitalOceanProvider, lego.DNS01},
	DNSIMPLE:     ProviderFactory{makeDNSimpleProvider, lego.DNS01},
	DYN:          ProviderFactory{makeDynProvider, lego.DNS01},
	GANDI:        ProviderFactory{makeGandiProvider, lego.DNS01},
	NS1:          ProviderFactory{makeNS1Provider, lego.DNS01},
	OVH:          ProviderFactory{makeOvhProvider, lego.DNS01},
	ROUTE53:      ProviderFactory{makeRoute53Provider, lego.DNS01},
	VULTR:        ProviderFactory{makeVultrProvider, lego.DNS01},
	HTTP:         ProviderFactory{makeHTTPProvider, lego.HTTP01},
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

// returns a preconfigured Aurora lego.ChallengeProvider
func makeAuroraProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.AuroraUserId) == 0 {
		return nil, fmt.Errorf("Aurora User Id is not set")
	}

	if len(opts.AuroraKey) == 0 {
		return nil, fmt.Errorf("Aurora Key is not set")
	}

	endpoint := opts.AuroraEndpoint
	if len(endpoint) == 0 {
		endpoint = "https://api.auroradns.eu"
	}

	provider, err := auroradns.NewDNSProviderCredentials(endpoint, opts.AuroraUserId,
		opts.AuroraKey)

	if err != nil {
		return nil, err
	}

	return provider, nil
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
	if len(opts.AwsAccessKey) != 0 {
		os.Setenv("AWS_ACCESS_KEY_ID", opts.AwsAccessKey)
	}
	if len(opts.AwsSecretKey) != 0 {
		os.Setenv("AWS_SECRET_ACCESS_KEY", opts.AwsSecretKey)
	}

	os.Setenv("AWS_REGION", "us-east-1")

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

// returns a preconfigured HTTP lego.ChallengeProvider
func makeHTTPProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	provider := lego.NewHTTPProviderServer("", "")
	return provider, nil
}

// returns a preconfigured Azure lego.ChallengeProvider
func makeAzureProvider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.AzureClientId) == 0 {
		return nil, fmt.Errorf("Azure Client ID is not set")
	}
	if len(opts.AzureClientSecret) == 0 {
		return nil, fmt.Errorf("Azure Client Secret is not set")
	}
	if len(opts.AzureSubscriptionId) == 0 {
		return nil, fmt.Errorf("Azure Subscription ID is not set")
	}
	if len(opts.AzureTenantId) == 0 {
		return nil, fmt.Errorf("Azure Tenant ID is not set")
	}
	if len(opts.AzureResourceGroup) == 0 {
		return nil, fmt.Errorf("Azure Resource Group is not set")
	}

	provider, err := azure.NewDNSProviderCredentials(opts.AzureClientId, opts.AzureClientSecret, opts.AzureSubscriptionId,
		opts.AzureTenantId, opts.AzureResourceGroup)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// returns a preconfigured NS1 lego.ChallengeProvider
func makeNS1Provider(opts ProviderOpts) (lego.ChallengeProvider, error) {
	if len(opts.NS1ApiKey) == 0 {
		return nil, fmt.Errorf("NS1 API key is not set")
	}

	provider, err := ns1.NewDNSProviderCredentials(opts.NS1ApiKey)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
