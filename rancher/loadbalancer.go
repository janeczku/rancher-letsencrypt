package rancher

import (
	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancher/go-rancher/client"
)

// UpgradeLoadBalancers upgrades all load balancers with the renewed certificate
func (r *Client) UpgradeLoadBalancers(certId string) error {
	balancers, err := r.findLoadBalancerServicesByCert(certId)
	if err != nil {
		return err
	}

	if len(balancers) == 0 {
		logrus.Info("Certificate is not being used by any load balancers")
		return nil
	}

	for _, id := range balancers {
		lb, err := r.client.LoadBalancerService.ById(id)
		if err != nil {
			logrus.Errorf("Failed to acquire load balancer by id %s: %v", id, err)
			continue
		}
		logrus.Infof("Upgrading load balancer %s...", lb.Name)
		err = r.upgrade(lb)
		if err != nil {
			logrus.Errorf("Failed to upgrade load balancer %s: %v", lb.Name, err)
		} else {
			logrus.Infof("Successfully upgraded load balancer %s with renewed certificate", lb.Name)
		}
	}

	return nil
}

func (r *Client) upgrade(lb *rancherClient.LoadBalancerService) error {
	upgrade := &rancherClient.ServiceUpgrade{}
	upgrade.InServiceStrategy = &rancherClient.InServiceUpgradeStrategy{
		LaunchConfig: lb.LaunchConfig,
		StartFirst:   false,
	}
	upgrade.ToServiceStrategy = &rancherClient.ToServiceUpgradeStrategy{}

	service, err := r.client.LoadBalancerService.ActionUpgrade(lb, upgrade)
	if err != nil {
		return err
	}

	err = r.WaitService(service)
	if err != nil {
		logrus.Warnf("Upgrade check: %v", err)
	}

	if service.State == "upgraded" {
		logrus.Debugf("Finishing upgrade of load balancer: %s", lb.Name)

		service, err = r.client.Service.ActionFinishupgrade(service)
		if err != nil {
			return err
		}
		err = r.WaitService(service)
		if err != nil {
			logrus.Warnf("Upgrade check: %v", err)
		}
	}

	return nil
}

func (r *Client) findLoadBalancerServicesByCert(certId string) ([]string, error) {
	var results []string

	logrus.Debugf("Looking up load balancers matching certificate id %s", certId)

	balancers, err := r.client.LoadBalancerService.List(&rancherClient.ListOpts{
		Filters: map[string]interface{}{
			"removed_null": nil,
			"state":        "active",
		},
	})
	if err != nil {
		return results, err
	}
	if len(balancers.Data) == 0 {
		logrus.Debug("Did not find matching load balancers")
		return results, nil
	}

	for _, b := range balancers.Data {
		if b.DefaultCertificateId == certId {
			results = append(results, b.Id)
			continue
		}
		for _, id := range b.CertificateIds {
			if id == certId {
				results = append(results, b.Id)
				break
			}
		}
	}

	logrus.Debugf("Found %d matching load balancers", len(results))
	return results, nil
}
