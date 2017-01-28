package rancher

import (
	"github.com/Sirupsen/logrus"
	rancherClient "github.com/rancher/go-rancher/v2"
)

// UpdateLoadBalancers updates all load balancers with the renewed certificate
func (r *Client) UpdateLoadBalancers(certId string) error {
	balancers, err := r.findLoadBalancerServicesByCert(certId)
	if err != nil {
		return err
	}

	if len(balancers) == 0 {
		logrus.Info("Certificate not used by any load balancer")
		return nil
	}

	for _, id := range balancers {
		lb, err := r.client.LoadBalancerService.ById(id)
		if err != nil {
			logrus.Errorf("Failed to get load balancer by ID %s: %v", id, err)
			continue
		}

		err = r.update(lb)
		if err != nil {
			logrus.Errorf("Failed to update load balancer '%s': %v", lb.Name, err)
		} else {
			logrus.Infof("Updated load balancer '%s' with changed certificate", lb.Name)
		}
	}

	return nil
}

func (r *Client) update(lb *rancherClient.LoadBalancerService) error {

	logrus.Debugf("Updating load balancer %s", lb.Name)

	service, err := r.client.LoadBalancerService.ActionUpdate(lb)
	if err != nil {
		return err
	}

	err = r.WaitService(service)
	if err != nil {
		logrus.Warnf(err.Error())
	}

	return nil
}

func (r *Client) findLoadBalancerServicesByCert(certId string) ([]string, error) {
	var results []string

	logrus.Debugf("Looking up load balancers matching certificate ID %s", certId)

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
		logrus.Debug("Did not find any active load balancers")
		return results, nil
	}

	logrus.Debugf("Found %d active load balancers", len(balancers.Data))

	for _, b := range balancers.Data {
		if b.LbConfig.DefaultCertificateId == certId {
			results = append(results, b.Id)
			continue
		}
		for _, id := range b.LbConfig.CertificateIds {
			if id == certId {
				results = append(results, b.Id)
				break
			}
		}
	}

	logrus.Debugf("Found %d load balancers with matching certificate", len(results))
	return results, nil
}
