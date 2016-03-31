package letsencrypt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	lego "github.com/xenolf/lego/acme"
)

const (
	ConfigDir        = "/etc/letsencrypt"
	ProductionApiUri = "https://acme-v01.api.letsencrypt.org/directory"
	StagingApiUri    = "https://acme-staging.api.letsencrypt.org/directory"
)

type KeyType string

const (
	RSA2048 KeyType = "RSA-2048"
	RSA4096 KeyType = "RSA-4096"
	RSA8192 KeyType = "RSA-8192"
	EC256   KeyType = "ECDSA-256"
	EC384   KeyType = "ECDSA-384"
)

type ApiVersion string

const (
	Production ApiVersion = "Production"
	Sandbox    ApiVersion = "Sandbox"
)

// AcmeCertificate represents a CA issued certificate,
// PrivateKey and Certificate are both PEM encoded.
//
// Anonymous fields:
// PrivateKey  []byte
// Certificate []byte
// Domain      string
type AcmeCertificate struct {
	lego.CertificateResource
	ExpiryDate   time.Time `json:"expiryDate"`
	SerialNumber string    `json:"serialnumber"`
}

// Client represents a Lets Encrypt client
type Client struct {
	client     *lego.Client
	apiVersion ApiVersion
}

// NewClient returns a new Lets Encrypt client
func NewClient(email string, kt KeyType, apiVer ApiVersion, provider ProviderOpts) (*Client, error) {
	var keyType lego.KeyType
	switch kt {
	case RSA2048:
		keyType = lego.RSA2048
	case RSA4096:
		keyType = lego.RSA4096
	case RSA8192:
		keyType = lego.RSA8192
	case EC256:
		keyType = lego.EC256
	case EC384:
		keyType = lego.EC384
	default:
		return nil, fmt.Errorf("Invalid private key type: %s", string(kt))
	}

	var serverUri string
	switch apiVer {
	case Production:
		serverUri = ProductionApiUri
	case Sandbox:
		serverUri = StagingApiUri
	default:
		return nil, fmt.Errorf("Invalid LE API version: %s", string(apiVer))
	}

	acc, err := NewAccount(email, apiVer, keyType)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize account store for %s: %v", email, err)
	}

	client, err := lego.NewClient(serverUri, acc, keyType)
	if err != nil {
		return nil, fmt.Errorf("Could not create client: %v", err)
	}

	lego.Logger = log.New(ioutil.Discard, "", 0)

	if acc.Registration == nil {
		logrus.Infof("Creating Let's Encrypt account for %s", email)
		reg, err := client.Register()
		if err != nil {
			return nil, fmt.Errorf("Failed to register account: %v", err)
		}

		acc.Registration = reg
		if acc.Registration.Body.Agreement == "" {
			err = client.AgreeToTOS()
			if err != nil {
				return nil, fmt.Errorf("Could not agree to TOS: %v", err)
			}
		}

		err = acc.Save()
		if err != nil {
			logrus.Errorf("Could not save account data: %v", err)
		}
	} else {
		logrus.Infof("Using locally stored Let's Encrypt account for %s", email)
	}

	prov, err := getProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("Could not set DNS provider: %v", err)
	}

	err = client.SetChallengeProvider(lego.DNS01, prov)
	if err != nil {
		return nil, fmt.Errorf("Could not set DNS provider: %v", err)
	}

	client.ExcludeChallenges([]lego.Challenge{lego.HTTP01, lego.TLSSNI01})

	return &Client{
		client:     client,
		apiVersion: apiVer,
	}, nil
}

// EnableDebugLogging enables logging in the upstream lego library
func (c *Client) EnableDebug() {
	lego.Logger = log.New(os.Stdout, "", 0)
}

// Issue obtains a new SAN certificate from the Lets Encrypt CA
func (c *Client) Issue(domains []string) (*AcmeCertificate, map[string]error) {
	cert, failures := c.client.ObtainCertificate(domains, true, nil)
	if len(failures) > 0 {
		return nil, failures
	}

	expirydate, _ := lego.GetPEMCertExpiration(cert.Certificate)
	serialnum, _ := getPEMCertSerialNo(cert.Certificate)

	acmeCert := AcmeCertificate{
		CertificateResource: cert,
		ExpiryDate:          expirydate,
		SerialNumber:        serialnum,
	}

	c.saveCertificate(acmeCert)
	return &acmeCert, nil
}

// Renew obtains a renewed SAN certificate from the Lets Encrypt CA
func (c *Client) Renew(domains []string) (*AcmeCertificate, error) {
	domain := domains[0]
	currentCert, err := c.loadCertificate(domain)
	if err != nil {
		logrus.Fatalf(err.Error())
	}
	certRes := currentCert.CertificateResource

	newCert, err := c.client.RenewCertificate(certRes, true)
	if err != nil {
		return nil, err
	}

	expirydate, _ := lego.GetPEMCertExpiration(newCert.Certificate)
	serialnum, _ := getPEMCertSerialNo(newCert.Certificate)

	acmeCert := AcmeCertificate{
		CertificateResource: newCert,
		ExpiryDate:          expirydate,
		SerialNumber:        serialnum,
	}

	c.saveCertificate(acmeCert)
	return &acmeCert, nil
}

// GetStoredCertificate returns a locally stored certificate for the given domains
func (c *Client) GetStoredCertificate(domains []string) (bool, *AcmeCertificate) {
	domain := domains[0]
	if !c.haveCertificate(domain) {
		return false, nil
	}

	acmeCert, err := c.loadCertificate(domain)
	if err != nil {
		// Don't fail here. Try to create a new certificate instead.
		logrus.Error(err.Error())
		return false, nil
	}
	return true, &acmeCert
}

func (c *Client) haveCertificate(domain string) bool {
	certPath := c.CertPath(domain)
	if _, err := os.Stat(path.Join(certPath, "metadata.json")); err != nil {
		logrus.Debugf("No existing acme certificate resource found for '%s'", domain)
		return false
	}
	return true
}

func (c *Client) loadCertificate(domain string) (AcmeCertificate, error) {
	var acmeCert AcmeCertificate
	certPath := c.CertPath(domain)

	logrus.Debugf("Loading acme certificate resource for '%s' from '%s'", domain, certPath)

	certIn := path.Join(certPath, "fullchain.pem")
	privIn := path.Join(certPath, "privkey.pem")
	metaIn := path.Join(certPath, "metadata.json")

	certBytes, err := ioutil.ReadFile(certIn)
	if err != nil {
		return acmeCert, fmt.Errorf("Failed to load certificate for domain '%s': %s", domain, err.Error())
	}

	metaBytes, err := ioutil.ReadFile(metaIn)
	if err != nil {
		return acmeCert, fmt.Errorf("Failed to load meta data for domain '%s': %s", domain, err.Error())
	}

	keyBytes, err := ioutil.ReadFile(privIn)
	if err != nil {
		return acmeCert, fmt.Errorf("Failed to load private key for domain '%s': %s", domain, err.Error())
	}

	err = json.Unmarshal(metaBytes, &acmeCert)
	if err != nil {
		return acmeCert, fmt.Errorf("Failed to unmarshal meta data for domain '%s': %s", domain, err.Error())
	}

	acmeCert.PrivateKey = keyBytes
	acmeCert.Certificate = certBytes

	return acmeCert, nil
}

func (c *Client) saveCertificate(acmeCert AcmeCertificate) {
	certPath := c.CertPath(acmeCert.Domain)
	maybeCreatePath(certPath)

	logrus.Debugf("Saving acme certificate resource for '%s' to '%s'", acmeCert.Domain, certPath)

	certOut := path.Join(certPath, "fullchain.pem")
	privOut := path.Join(certPath, "privkey.pem")
	metaOut := path.Join(certPath, "metadata.json")

	err := ioutil.WriteFile(certOut, acmeCert.Certificate, 0600)
	if err != nil {
		logrus.Fatalf("Failed to save certificate for domain %s\n\t%s", acmeCert.Domain, err.Error())
	}

	err = ioutil.WriteFile(privOut, acmeCert.PrivateKey, 0600)
	if err != nil {
		logrus.Fatalf("Failed to save private key for domain %s\n\t%s", acmeCert.Domain, err.Error())
	}

	jsonBytes, err := json.MarshalIndent(acmeCert, "", "\t")
	if err != nil {
		logrus.Fatalf("Failed to marshal meta data for domain %s\n\t%s", acmeCert.Domain, err.Error())
	}

	err = ioutil.WriteFile(metaOut, jsonBytes, 0600)
	if err != nil {
		logrus.Fatalf("Failed to save meta data for domain %s\n\t%s", acmeCert.Domain, err.Error())
	}
}

func (c *Client) ConfigPath() string {
	path := path.Join(ConfigDir, strings.ToLower(string(c.apiVersion)))
	maybeCreatePath(path)
	return path
}

func (c *Client) CertPath(domain string) string {
	return path.Join(c.ConfigPath(), "certificates", domain)
}

func maybeCreatePath(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			logrus.Fatalf("Error creating path: %v", err)
		}
	}
}
