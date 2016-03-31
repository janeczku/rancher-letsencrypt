package main

import (
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	RENEW_BEFORE_DAYS = 14
)

func (c *Context) Run() {
	c.startup()
	for {
		<-c.timer()
		c.renew()
	}
}

func (c *Context) startup() {
	var haveLocal, haveRemote bool
	ok, acmeCert := c.Acme.GetStoredCertificate(c.Domains)
	if ok {
		haveLocal = true
		c.ExpiryDate = acmeCert.ExpiryDate
		logrus.Infof("Found locally stored certificate for '%s'", strings.Join(c.Domains, " | "))
	}

	rancherCert, err := c.Rancher.FindCertByName(c.RancherCertName)
	if err != nil {
		logrus.Fatalf("Error looking up Rancher certificates: %v", err)
	}

	if rancherCert != nil {
		haveRemote = true
		c.RancherCertId = rancherCert.Id
		logrus.Infof("Found existing Rancher certificate '%s'", rancherCert.Name)
	}

	if haveLocal && haveRemote {
		if rancherCert.SerialNumber == acmeCert.SerialNumber {
			logrus.Infof("Managing renewal of Rancher certificate '%s'", rancherCert.Name)
			return
		}
		logrus.Infof("Certificate serial number mismatch. Overwriting Rancher certificate '%s'", rancherCert.Name)
		c.updateRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	if haveLocal && !haveRemote {
		logrus.Infof("Adding Rancher certificate '%s'", rancherCert.Name)
		c.addRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	logrus.Infof("Trying to obtain SSL certificate for %s", strings.Join(c.Domains, " | "))

	acmeCert, failures := c.Acme.Issue(c.Domains)
	if len(failures) > 0 {
		for k, v := range failures {
			logrus.Errorf("[%s] Error obtaining certificate: %s", k, v.Error())
		}
		os.Exit(1)
	}

	logrus.Infof("Successfully obtained SSL certificate")

	c.ExpiryDate = acmeCert.ExpiryDate

	if haveRemote {
		logrus.Infof("Overwriting Rancher certificate '%s'", rancherCert.Name)
		c.updateRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	c.addRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
}

func (c *Context) addRancherCert(privateKey, cert []byte) {
	rancherCert, err := c.Rancher.AddCertificate(c.RancherCertName, CERT_DESCRIPTION, privateKey, cert)
	if err != nil {
		logrus.Fatalf("Failed to add Rancher certificate '%s': %v", c.RancherCertName, err)
	}
	c.RancherCertId = rancherCert.Id
	logrus.Infof("Added Rancher certificate '%s'", c.RancherCertName)
}

func (c *Context) updateRancherCert(privateKey, cert []byte) {
	err := c.Rancher.UpdateCertificate(c.RancherCertId, CERT_DESCRIPTION, privateKey, cert)
	if err != nil {
		logrus.Fatalf("Failed to update Rancher certificate '%s': %v", c.RancherCertName, err)
	}
	logrus.Infof("Updated Rancher certificate '%s'", c.RancherCertName)

	err = c.Rancher.UpgradeLoadBalancers(c.RancherCertId)
	if err != nil {
		logrus.Fatalf("Error upgrading load balancers: %v", err)
	}
}

func (c *Context) renew() {
	logrus.Infof("Trying to renew certificate for '%s'", strings.Join(c.Domains, " | "))

	acmeCert, err := c.Acme.Renew(c.Domains)
	if err != nil {
		logrus.Fatalf("Failed to renew certificate: %v", err)
	}

	logrus.Infof("Successfully renewed certificate")

	c.ExpiryDate = acmeCert.ExpiryDate
	c.updateRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
}

func (c *Context) timer() <-chan time.Time {
	now := time.Now().UTC()
	next := c.getRenewalDate()
	left := next.Sub(now)
	if left <= 0 {
		left = 10 * time.Second
	}

	logrus.Infof("Next certificate renewal scheduled for %s", next.Format("2006/01/02 15:04 MST"))

	// Debug option set to true enables test mode
	if c.Debug {
		logrus.Debug("Test mode enabled: certificate will be renewed in 120 seconds")
		left = 120 * time.Second
	}

	return time.After(left)
}

func (c *Context) getRenewalDate() time.Time {
	date := c.ExpiryDate.AddDate(0, 0, -RENEW_BEFORE_DAYS)
	dYear, dMonth, dDay := date.Date()
	return time.Date(dYear, dMonth, dDay, c.RenewalTime, 0, 0, 0, time.UTC)
}
