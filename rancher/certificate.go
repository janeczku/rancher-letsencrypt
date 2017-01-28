package rancher

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancher/go-rancher/v2"
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

	logrus.Debugf("Waiting for new certificate '%s' to become active", rancherCert.Name)

	if err := r.WaitCertificate(rancherCert); err != nil {
		return nil, err
	}

	return rancherCert, nil
}

// UpdateCertificate updates an existing certificate resource using the given PEM encoded certificate
func (r *Client) UpdateCertificate(certId, descr string, privateKey, cert []byte) error {
	certString := string(cert[:])
	keyString := string(privateKey[:])
	rancherCert, err := r.client.Certificate.ById(certId)
	if err != nil {
		return err
	}

	rancherCert, err = r.client.Certificate.Update(rancherCert, &rancherClient.Certificate{
		Description: descr,
		Cert:        certString,
		Key:         keyString,
	})
	if err != nil {
		return err
	}

	logrus.Debugf("Waiting for updated certificate '%s' to become active", rancherCert.Name)

	return r.WaitCertificate(rancherCert)
}

// FindCertByName retrieves an existing certificate
func (r *Client) FindCertByName(name string) (*rancherClient.Certificate, error) {
	logrus.Debugf("Looking up Rancher certificate by name: %s", name)

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

	logrus.Debugf("Found existing Rancher certificate by name: %s", name)
	return &certificates.Data[0], nil
}

// GetCertById retrieves an existing certificate by ID
func (r *Client) GetCertById(certId string) (*rancherClient.Certificate, error) {
	rancherCert, err := r.client.Certificate.ById(certId)
	if err != nil {
		return nil, err
	}

	if rancherCert == nil {
		return nil, fmt.Errorf("No such certificate with ID %s", certId)
	}

	logrus.Debugf("Got Rancher certificate %s by ID %s", rancherCert.Name, certId)
	return rancherCert, nil
}
