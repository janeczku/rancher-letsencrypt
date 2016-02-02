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
	DESCRIPTION       = "Created by Let's Encrypt Certificate Manager"
	ISSUER_PRODUCTION = "Let's Encrypt"
	ISSUER_STAGING    = "fake CA"
	PREFIX_PROD       = "[LE] "
	PREFIX_STAGING    = "[LE-TESTING] "
)

type Context struct {
	Acme    *letsencrypt.Client
	Rancher *rancher.Client

	Domains     []string
	RenewalTime int

	ExpiryDate      time.Time
	RancherCertId   string
	RancherCertName string

	Debug bool
}

// InitContext initializes the application context from environmental variables
func (c *Context) InitContext() {
	cattleUrl := getEnvOption("CATTLE_URL", true)
	cattleApiKey := getEnvOption("CATTLE_ACCESS_KEY", true)
	cattleSecretKey := getEnvOption("CATTLE_SECRET_KEY", true)
	debug := getEnvOption("DEBUG", false)
	eulaParam := getEnvOption("EULA", false)
	apiVerParam := getEnvOption("API_VERSION", true)
	emailParam := getEnvOption("EMAIL", true)
	domainParam := getEnvOption("DOMAINS", true)
	keyTypeParam := getEnvOption("PUBLIC_KEY_TYPE", true)
	timeParam := getEnvOption("RENEWAL_TIME", true)
	providerParam := getEnvOption("PROVIDER", true)

	if eulaParam != "Yes" {
		logrus.Fatalf("Terms of service were not accepted")
	}

	var err error

	c.Domains = listToSlice(domainParam)
	if len(c.Domains) == 0 {
		logrus.Fatalf("Invalid value for DOMAINS: %s", domainParam)
	}

	c.RenewalTime, err = strconv.Atoi(timeParam)
	if err != nil || c.RenewalTime < 0 || c.RenewalTime > 23 {
		logrus.Fatalf("Invalid value for RENEWAL_TIME: %s", timeParam)
	}

	var serverURI string
	switch apiVerParam {
	case "Production":
		serverURI = letsencrypt.PRODUCTION_URI
		c.RancherCertName = PREFIX_PROD + c.Domains[0]
	case "Sandbox":
		serverURI = letsencrypt.STAGING_URI
		c.RancherCertName = PREFIX_STAGING + c.Domains[0]
	default:
		logrus.Fatalf("Invalid value for API_VERSION: %s", apiVerParam)
	}

	c.Rancher, err = rancher.NewClient(cattleUrl, cattleApiKey, cattleSecretKey)
	if err != nil {
		logrus.Fatalf("Rancher client: %v", err)
	}

	providerOpts := letsencrypt.ProviderOpts{
		Provider:        letsencrypt.DnsProvider(providerParam),
		CloudflareEmail: os.Getenv("CLOUDFLARE_EMAIL"),
		CloudflareKey:   os.Getenv("CLOUDFLARE_KEY"),
		DoAccessToken:   os.Getenv("DO_ACCESS_TOKEN"),
		AwsAccessKey:    os.Getenv("AWS_ACCESS_KEY"),
		AwsSecretKey:    os.Getenv("AWS_SECRET_KEY"),
		AwsRegionName:   "us-east-1",
	}

	c.Acme, err = letsencrypt.NewClient(emailParam, keyTypeParam, serverURI, providerOpts)
	if err != nil {
		logrus.Fatalf("LetsEncrypt client: %v", err)
	}

	// Enable debug/test mode
	if strings.EqualFold(debug, "true") {
		logrus.SetLevel(logrus.DebugLevel)
		c.Debug = true
		c.Acme.EnableDebugLogging()
	}
}

func getEnvOption(name string, required bool) string {
	val := os.Getenv(name)
	if required && len(val) == 0 {
		logrus.Fatalf("Required environment variable not set: %s", name)
	}
	return val
}

func listToSlice(str string) []string {
	str = strings.Join(strings.Fields(str), "")
	return strings.Split(str, ",")
}
