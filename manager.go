package main

import (
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

func (c *Context) Run() {
	c.startup()
	if c.RunOnce {
		// Renew certificate if it's about to expire
		if time.Now().UTC().After(c.getRenewalDate()) {
			c.renew()
		} else {
			logrus.Infof("Not renewing certificate %s which expires on %s", c.CertificateName,
				c.ExpiryDate.UTC().Format(time.UnixDate))
		}
		logrus.Info("Run once: Finished")
		return
	}

	for {
		<-c.timer()
		c.renew()
	}
}

func (c *Context) startup() {
	var storedLocally, storedInRancher bool
	ok, acmeCert := c.Acme.GetStoredCertificate(c.CertificateName, c.Domains)
	if ok {
		storedLocally = true
		c.ExpiryDate = acmeCert.ExpiryDate
		logrus.Infof("Found locally stored certificate '%s'", c.CertificateName)
	}

	rancherCert, err := c.Rancher.FindCertByName(c.CertificateName)
	if err != nil {
		logrus.Fatalf("Could not lookup certificate in Rancher API: %v", err)
	}

	if rancherCert != nil {
		storedInRancher = true
		c.RancherCertId = rancherCert.Id
		logrus.Infof("Found existing certificate '%s' in Rancher", c.CertificateName)
	}

	if storedLocally && storedInRancher {
		if rancherCert.SerialNumber == acmeCert.SerialNumber {
			logrus.Infof("Managing renewal of certificate '%s'", c.CertificateName)
			return
		}
		logrus.Infof("Serial number mismatch between Rancher and local certificate '%s'", c.CertificateName)
		c.updateRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	if storedLocally && !storedInRancher {
		logrus.Debugf("Adding certificate '%s' to Rancher", c.CertificateName)
		c.addRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	if c.Acme.ProviderName() == "HTTP" {
		logrus.Info("Using HTTP challenge: Sleeping for 120 seconds before requesting certificate")
		logrus.Info("Make sure that HTTP requests for '/.well-known/acme-challenge' for all certificate " +
			"domains are forwarded to port 80 of the container running this application")
		time.Sleep(120 * time.Second)
	}

	logrus.Infof("Trying to obtain SSL certificate (%s) from Let's Encrypt %s CA", strings.Join(c.Domains, ","), c.Acme.ApiVersion())

	acmeCert, err = c.Acme.Issue(c.CertificateName, c.Domains)
	if err != nil {
		logrus.Errorf("[%s] Error obtaining certificate: %s", err, err.Error())
		os.Exit(1)
	}

	logrus.Infof("Certificate obtained successfully")

	c.ExpiryDate = acmeCert.ExpiryDate

	if storedInRancher {
		logrus.Debugf("Overwriting Rancher certificate '%s'", c.CertificateName)
		c.updateRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	c.addRancherCert(acmeCert.PrivateKey, acmeCert.Certificate)
}

func (c *Context) addRancherCert(privateKey, cert []byte) {
	rancherCert, err := c.Rancher.AddCertificate(c.CertificateName, CERT_DESCRIPTION, privateKey, cert)
	if err != nil {
		logrus.Fatalf("Failed to add Rancher certificate '%s': %v", c.CertificateName, err)
	}
	c.RancherCertId = rancherCert.Id
	logrus.Infof("Certificate '%s' added to Rancher", c.CertificateName)
}

func (c *Context) updateRancherCert(privateKey, cert []byte) {
	err := c.Rancher.UpdateCertificate(c.RancherCertId, CERT_DESCRIPTION, privateKey, cert)
	if err != nil {
		logrus.Fatalf("Failed to update Rancher certificate '%s': %v", c.CertificateName, err)
	}
	logrus.Infof("Updated Rancher certificate '%s'", c.CertificateName)

	err = c.Rancher.UpdateLoadBalancers(c.RancherCertId)
	if err != nil {
		logrus.Fatalf("Failed to upgrade load balancers: %v", err)
	}
}

func (c *Context) renew() {
	logrus.Infof("Trying to obtain renewed SSL certificate (%s) from Let's Encrypt %s CA", strings.Join(c.Domains, ","), c.Acme.ApiVersion())

	acmeCert, err := c.Acme.Renew(c.CertificateName)
	if err != nil {
		logrus.Fatalf("Failed to renew certificate: %v", err)
	}

	logrus.Infof("Certificate renewed successfully")

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

	logrus.Infof("Certificate renewal scheduled for %s", next.Format("2006/01/02 15:04 MST"))

	// test mode forces renewal
	if c.TestMode {
		logrus.Debug("Test mode: Forced certificate renewal in 120 seconds")
		left = 120 * time.Second
	}

	return time.After(left)
}

func (c *Context) getRenewalDate() time.Time {
	if c.ExpiryDate.IsZero() {
		logrus.Fatalf("Could not determine expiry date for certificate: %s", c.CertificateName)
	}
	date := c.ExpiryDate.AddDate(0, 0, -c.RenewalPeriodDays)
	dYear, dMonth, dDay := date.Date()
	return time.Date(dYear, dMonth, dDay, c.RenewalDayTime, 0, 0, 0, time.UTC)
}
