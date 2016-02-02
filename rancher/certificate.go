package rancher

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancher/go-rancher/client"
)

// AddCertificate creates a new certificate resource using the given private key and PEM encoded certificate
func (r *Client) AddCertificate(name, descr string, privateKey, cert []byte) (*rancherClient.Certificate, error) {
	certString := string(cert[:])
	keyString := string(privateKey[:])

	config := &rancherClient.Certificate{
		Name:        name,
		Description: descr,
		Cert:        certString,
		Key:         keyString,
	}

	rancherCert, err := r.client.Certificate.Create(config)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Added certificate: %s", rancherCert.Name)

	if err := r.WaitCertificate(rancherCert); err != nil {
		return nil, err
	}

	return rancherCert, nil
}

// UpdateCertificate updates an existing certificate resource using the given PEM encoded certificate
func (r *Client) UpdateCertificate(certId string, cert []byte) error {
	certString := string(cert[:])
	rancherCert, err := r.client.Certificate.ById(certId)
	if err != nil {
		return err
	}

	rancherCert, err = r.client.Certificate.Update(rancherCert, &rancherClient.Certificate{
		Cert: certString,
	})
	if err != nil {
		return err
	}

	logrus.Debugf("Updated certificate %s", rancherCert.Name)

	return r.WaitCertificate(rancherCert)
}

// FindCertByName retrieves an existing certificate
func (r *Client) FindCertByName(name string) (*rancherClient.Certificate, error) {
	logrus.Debugf("Looking up certificate by name %s", name)

	certificates, err := r.client.Certificate.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"name":         name,
			"removed_null": nil,
		},
	})

	if err != nil {
		return nil, err
	}

	if len(certificates.Data) == 0 {
		return nil, nil
	}

	logrus.Debugf("Found existing certificate %s", name)
	return &certificates.Data[0], nil
}

// GetCertById retrieves an existing certificate by ID
func (r *Client) GetCertById(certId string) (*rancherClient.Certificate, error) {
	rancherCert, err := r.client.Certificate.ById(certId)
	if err != nil {
		return nil, err
	}

	if rancherCert == nil {
		return nil, fmt.Errorf("Certificate with Id %s does not exist.", certId)
	}

	logrus.Debugf("Found certificate %s by Id %s", rancherCert.Name, certId)
	return rancherCert, nil
}
