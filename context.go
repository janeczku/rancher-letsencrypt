package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/janeczku/rancher-letsencrypt/letsencrypt"
	"github.com/janeczku/rancher-letsencrypt/rancher"
)

const (
	CERT_DESCRIPTION  = "Created by Let's Encrypt Certificate Manager"
	ISSUER_PRODUCTION = "Let's Encrypt"
	ISSUER_STAGING    = "fake CA"
)

type Context struct {
	Acme    *letsencrypt.Client
	Rancher *rancher.Client

	CertificateName string
	Domains         []string
	RenewalTime     int

	ExpiryDate    time.Time
	RancherCertId string

	Debug    bool
	TestMode bool
}

// InitContext initializes the application context from environmental variables
func (c *Context) InitContext() {
	var err error
	c.Debug = debug
	c.TestMode = testMode
	cattleUrl := getEnvOption("CATTLE_URL", true)
	cattleApiKey := getEnvOption("CATTLE_ACCESS_KEY", true)
	cattleSecretKey := getEnvOption("CATTLE_SECRET_KEY", true)
	eulaParam := getEnvOption("EULA", false)
	apiVerParam := getEnvOption("API_VERSION", true)
	emailParam := getEnvOption("EMAIL", true)
	domainParam := getEnvOption("DOMAINS", true)
	keyTypeParam := getEnvOption("PUBLIC_KEY_TYPE", true)
	certNameParam := getEnvOption("CERT_NAME", true)
	timeParam := getEnvOption("RENEWAL_TIME", true)
	providerParam := getEnvOption("PROVIDER", true)

	if eulaParam != "Yes" {
		logrus.Fatalf("Terms of service were not accepted")
	}

	c.Domains = listToSlice(domainParam)
	if len(c.Domains) == 0 {
		logrus.Fatalf("Invalid value for DOMAINS: %s", domainParam)
	}

	c.CertificateName = certNameParam
	c.RenewalTime, err = strconv.Atoi(timeParam)
	if err != nil || c.RenewalTime < 0 || c.RenewalTime > 23 {
		logrus.Fatalf("Invalid value for RENEWAL_TIME: %s", timeParam)
	}

	apiVersion := letsencrypt.ApiVersion(apiVerParam)
	keyType := letsencrypt.KeyType(keyTypeParam)

	c.Rancher, err = rancher.NewClient(cattleUrl, cattleApiKey, cattleSecretKey)
	if err != nil {
		logrus.Fatalf("Could not connect to Rancher API: %v", err)
	}

	providerOpts := letsencrypt.ProviderOpts{
		Provider:             letsencrypt.Provider(providerParam),
		AzureClientId:        getEnvOption("AZURE_CLIENT_ID", false),
		AzureClientSecret:    getEnvOption("AZURE_CLIENT_SECRET", false),
		AzureSubscriptionId:  getEnvOption("AZURE_SUBSCRIPTION_ID", false),
		AzureTenantId:        getEnvOption("AZURE_TENANT_ID", false),
		AzureResourceGroup:   getEnvOption("AZURE_RESOURCE_GROUP", false),
		AuroraUserId:         getEnvOption("AURORA_USER_ID", false),
		AuroraKey:            getEnvOption("AURORA_KEY", false),
		AuroraEndpoint:       getEnvOption("AURORA_ENDPOINT", false),
		CloudflareEmail:      getEnvOption("CLOUDFLARE_EMAIL", false),
		CloudflareKey:        getEnvOption("CLOUDFLARE_KEY", false),
		DoAccessToken:        getEnvOption("DO_ACCESS_TOKEN", false),
		GoogleCloudProject:   getEnvOption("GCE_PROJECT", false),
		AwsAccessKey:         getEnvOption("AWS_ACCESS_KEY", false),
		AwsSecretKey:         getEnvOption("AWS_SECRET_KEY", false),
		DNSimpleEmail:        getEnvOption("DNSIMPLE_EMAIL", false),
		DNSimpleKey:          getEnvOption("DNSIMPLE_KEY", false),
		DynCustomerName:      getEnvOption("DYN_CUSTOMER_NAME", false),
		DynUserName:          getEnvOption("DYN_USER_NAME", false),
		DynPassword:          getEnvOption("DYN_PASSWORD", false),
		VultrApiKey:          getEnvOption("VULTR_API_KEY", false),
		OvhApplicationKey:    getEnvOption("OVH_APPLICATION_KEY", false),
		OvhApplicationSecret: getEnvOption("OVH_APPLICATION_SECRET", false),
		OvhConsumerKey:       getEnvOption("OVH_CONSUMER_KEY", false),
		GandiApiKey:          getEnvOption("GANDI_API_KEY", false),
		NS1ApiKey:            getEnvOption("NS1_API_KEY", false),
	}

	c.Acme, err = letsencrypt.NewClient(emailParam, keyType, apiVersion, providerOpts)
	if err != nil {
		logrus.Fatalf("LetsEncrypt client: %v", err)
	}

	logrus.Infof("Using Let's Encrypt %s API", apiVersion)

	// Enable debug mode
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		c.Acme.EnableDebug()
	}
}

func getEnvOption(name string, required bool) string {
	val := os.Getenv(name)
	if required && len(val) == 0 {
		logrus.Fatalf("Required environment variable not set: %s", name)
	}
	return strings.TrimSpace(val)
}

func listToSlice(str string) []string {
	str = strings.ToLower(strings.Join(strings.Fields(str), ""))
	return strings.Split(str, ",")
}
