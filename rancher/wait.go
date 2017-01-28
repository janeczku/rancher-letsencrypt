package rancher

import (
	"errors"
	"fmt"
	"time"

	rancherClient "github.com/rancher/go-rancher/v2"
)

func backoff(maxDuration time.Duration, timeoutMessage string, f func() (bool, error)) error {
	startTime := time.Now()
	waitTime := 150 * time.Millisecond
	maxWaitTime := 2 * time.Second
	for {
		if time.Now().Sub(startTime) > maxDuration {
			return errors.New(timeoutMessage)
		}

		if done, err := f(); err != nil {
			return err
		} else if done {
			return nil
		}

		time.Sleep(waitTime)

		waitTime *= 2
		if waitTime > maxWaitTime {
			waitTime = maxWaitTime
		}
	}
}

// WaitFor waits for a resource to reach a certain state.
func (r *Client) WaitFor(resource *rancherClient.Resource, output interface{}, transitioning func() string) error {
	return backoff(2*time.Minute, fmt.Sprintf("Time out waiting for %s:%s to become active", resource.Type, resource.Id), func() (bool, error) {
		err := r.client.Reload(resource, output)
		if err != nil {
			return false, err
		}
		if transitioning() != "yes" {
			return true, nil
		}
		return false, nil
	})
}

//  WaitService waits for a loadbalancer resource to transition
func (r *Client) WaitService(service *rancherClient.Service) error {
	return r.WaitFor(&service.Resource, service, func() string {
		return service.Transitioning
	})
}

// WaitLoadBalancerService waits for a loadbalancer service resource to transition
func (r *Client) WaitLoadBalancerService(lb *rancherClient.LoadBalancerService) error {
	return r.WaitFor(&lb.Resource, lb, func() string {
		return lb.Transitioning
	})
}

// WaitCertificate waits for a certificate resource to transition
func (r *Client) WaitCertificate(certificate *rancherClient.Certificate) error {
	return r.WaitFor(&certificate.Resource, certificate, func() string {
		return certificate.Transitioning
	})
}
