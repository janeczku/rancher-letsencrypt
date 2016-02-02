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
	var haveLocal bool
	var haveRemote bool

	ok, acmeCert := c.Acme.GetStoredCert(c.Domains)
	if ok {
		haveLocal = true
		c.ExpiryDate = acmeCert.ExpiryDate
		logrus.Info("Found local store for certificate")
	}

	rancherCert, err := c.Rancher.FindCertByName(c.RancherCertName)
	if err != nil {
		logrus.Fatalf("Failed to lookup Rancher certificates: %v", err)
	}

	if rancherCert != nil {
		haveRemote = true
		c.RancherCertId = rancherCert.Id
	}

	if haveLocal && haveRemote {
		if rancherCert.SerialNumber != acmeCert.SerialNumber {
			logrus.Fatalf("Cannot manage existing Rancher certificate %s: Serial number not matching local store", rancherCert.Name)
		}
		logrus.Infof("Managing existing Rancher certificate: %s", rancherCert.Name)
		return
	}

	if !haveLocal && haveRemote {
		logrus.Fatalf("Cannot manage existing Rancher certificate %s: Not in local store", rancherCert.Name)
	}

	if haveLocal && !haveRemote {
		c.addCertToRancher(acmeCert.PrivateKey, acmeCert.Certificate)
		return
	}

	logrus.Infof("Trying to obtain SSL certificate for %s", strings.Join(c.Domains, " | "))

	acmeCert, failures := c.Acme.Issue(c.Domains)
	if len(failures) > 0 {
		for k, v := range failures {
			logrus.Errorf("[%s] Failed to obtain certificate: %s", k, v.Error())
		}
		os.Exit(1)
	}

	logrus.Infof("Successfully obtained certificate")

	c.ExpiryDate = acmeCert.ExpiryDate

	c.addCertToRancher(acmeCert.PrivateKey, acmeCert.Certificate)
}

func (c *Context) addCertToRancher(privateKey, cert []byte) {
	rancherCert, err := c.Rancher.AddCertificate(c.RancherCertName, DESCRIPTION, privateKey, cert)
	if err != nil {
		logrus.Fatalf("Failed to add Rancher certificate resource: %v", err)
	}
	c.RancherCertId = rancherCert.Id
	logrus.Infof("Added Rancher certificate resource: %s", c.RancherCertName)
}

func (c *Context) renew() {
	logrus.Infof("Trying to renew certificate for %s", strings.Join(c.Domains, " | "))

	acmeCert, err := c.Acme.Renew(c.Domains)
	if err != nil {
		logrus.Fatalf("Failed to renew certificate: %v", err)
	}

	logrus.Infof("Successfully renewed certificate")

	c.ExpiryDate = acmeCert.ExpiryDate
	err = c.Rancher.UpdateCertificate(c.RancherCertId, acmeCert.Certificate)
	if err != nil {
		logrus.Fatalf("Failed to update Rancher certificate resource: %v", err)
	}

	logrus.Infof("Updated Rancher certificate resource %s", c.RancherCertName)

	err = c.Rancher.UpgradeLoadBalancers(c.RancherCertId)
	if err != nil {
		logrus.Fatalf("Error upgrading load balancers: %v", err)
	}
}

func (c *Context) timer() <-chan time.Time {
	now := time.Now().UTC()
	next := c.getRenewalDate()
	left := next.Sub(now)
	if left <= 0 {
		left = 10 * time.Second
	}

	// Test mode
	if c.Debug {
		logrus.Debug("Test mode enabled: Certificate renewal in 120 seconds")
		left = 120 * time.Second
	}

	logrus.Infof("Next certificate renewal scheduled for %s", next.Format("2006/01/02 15:04 MST"))
	return time.After(left)
}

func (c *Context) getRenewalDate() time.Time {
	date := c.ExpiryDate.AddDate(0, 0, -RENEW_BEFORE_DAYS)
	dYear, dMonth, dDay := date.Date()
	return time.Date(dYear, dMonth, dDay, c.RenewalTime, 0, 0, 0, time.UTC)
}
