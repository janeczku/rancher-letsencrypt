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
	PRODUCTION_URI = "https://acme-v01.api.letsencrypt.org/directory"
	STAGING_URI    = "https://acme-staging.api.letsencrypt.org/directory"
	DATA_DIR       = "/etc/letsencrypt"
)

// AcmeCertificate represents a CA issued certificate,
// PrivateKey and Certificate are both PEM encoded.
type AcmeCertificate struct {
	PrivateKey   []byte    `json:"privateKey"`
	Certificate  []byte    `json:"certificate"`
	ExpiryDate   time.Time `json:"expiryDate"`
	SerialNumber string    `json:"serialnumber"`
}

type acmeCertificateStore struct {
	CertRes  lego.CertificateResource `json:"certRes"`
	AcmeCert AcmeCertificate          `json:"acmeCert"`
}

// Client represents a Lets Encrypt client
type Client struct {
	client *lego.Client
}

// NewClient returns a new Lets Encrypt client
func NewClient(email, keyTypeStr, uri string, provider ProviderOpts) (*Client, error) {
	var keyType lego.KeyType
	switch strings.ToUpper(keyTypeStr) {
	case "RSA-2048":
		keyType = lego.RSA2048
	case "RSA-4096":
		keyType = lego.RSA4096
	case "RSA-8192":
		keyType = lego.RSA8192
	case "ECDSA-256":
		keyType = lego.EC256
	case "ECDSA-384":
		keyType = lego.EC384
	default:
		return nil, fmt.Errorf("Invalid private key type: %s", keyTypeStr)
	}

	acc, err := NewAccount(email, keyType)
	if err != nil {
		return nil, fmt.Errorf("Could not initialize account store for %s: %v", email, err)
	}

	client, err := lego.NewClient(uri, acc, keyType)
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
		client: client,
	}, nil
}

// Issue obtains a new SAN certificate from the Lets Encrypt CA
func (c *Client) Issue(domains []string) (*AcmeCertificate, map[string]error) {
	cert, failures := c.client.ObtainCertificate(domains, true, nil)
	if len(failures) > 0 {
		return nil, failures
	}

	expirydate, _ := lego.GetPEMCertExpiration(cert.Certificate)
	serial, _ := getPEMCertSerialNo(cert.Certificate)

	acmeCert := AcmeCertificate{
		cert.PrivateKey,
		cert.Certificate,
		expirydate,
		serial,
	}

	saveCertStore(cert, acmeCert)
	return &acmeCert, nil
}

// Renew obtains a renewed SAN certificate from the Lets Encrypt CA
func (c *Client) Renew(domains []string) (*AcmeCertificate, error) {
	domain := domains[0]
	certStore, err := loadCertStore(domain)
	if err != nil {
		return nil, err
	}

	certRes := certStore.CertRes
	certRes.PrivateKey = certStore.AcmeCert.PrivateKey
	certRes.Certificate = certStore.AcmeCert.Certificate

	newCert, err := c.client.RenewCertificate(certRes, true)
	if err != nil {
		return nil, err
	}

	expirydate, _ := lego.GetPEMCertExpiration(newCert.Certificate)
	serial, _ := getPEMCertSerialNo(newCert.Certificate)

	acmeCert := AcmeCertificate{
		newCert.PrivateKey,
		newCert.Certificate,
		expirydate,
		serial,
	}

	saveCertStore(newCert, acmeCert)
	return &acmeCert, nil
}

// GetStoredCert returns a locally stored certificate for the given domains
func (c *Client) GetStoredCert(domains []string) (bool, *AcmeCertificate) {
	domain := domains[0]
	if !haveCertStore(domain) {
		return false, nil
	}

	certStore, err := loadCertStore(domain)
	if err != nil {
		logrus.Error(err)
		return false, nil
	}
	return true, &certStore.AcmeCert
}

// EnableDebugLogging enables logging in the upstream lego library
func (c *Client) EnableDebugLogging() {
	lego.Logger = log.New(os.Stdout, "", 0)
}

func haveCertStore(domain string) bool {
	path := certStorePath(domain)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func loadCertStore(domain string) (*acmeCertificateStore, error) {
	file := certStorePath(domain)
	certStoreBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Error loading certificate store for %s: %v", domain, err)
	}

	var certStore acmeCertificateStore
	err = json.Unmarshal(certStoreBytes, &certStore)
	if err != nil {
		return nil, fmt.Errorf("Error parsing certificate store for %s: %v", domain, err)
	}

	return &certStore, nil
}

func saveCertStore(certRes lego.CertificateResource, acmeCert AcmeCertificate) {
	out := certStorePath(certRes.Domain)

	certStore := acmeCertificateStore{
		CertRes:  certRes,
		AcmeCert: acmeCert,
	}

	jsonBytes, err := json.MarshalIndent(certStore, "", "\t")
	if err != nil {
		logrus.Fatalf("Unable to marshal certificate store for domain %s: %s", certRes.Domain, err.Error())
	}

	err = ioutil.WriteFile(out, jsonBytes, 0600)
	if err != nil {
		logrus.Fatalf("Unable to save certificate store for domain %s: %s", certRes.Domain, err.Error())
	}
}

func configPath() string {
	path := path.Join(DATA_DIR, ".acme")
	if err := checkFolder(path); err != nil {
		logrus.Fatalf("Could not check/create config directory: %v", err)
	}
	return path
}

func certStorePath(domain string) string {
	path := path.Join(configPath(), domain+".json")
	return path
}

func checkFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700)
	}
	return nil
}
