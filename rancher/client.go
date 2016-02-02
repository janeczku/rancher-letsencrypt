package rancher

import rancherClient "github.com/rancher/go-rancher/client"

type Client struct {
	client *rancherClient.RancherClient
}

// NewClient returns a new client for the Rancher/Cattle API
func NewClient(rancherUrl string, rancherAccessKey string, rancherSecretKey string) (*Client, error) {
	apiClient, err := rancherClient.NewRancherClient(&rancherClient.ClientOpts{
		Url:       rancherUrl,
		AccessKey: rancherAccessKey,
		SecretKey: rancherSecretKey,
	})

	if err != nil {
		return nil, err
	}

	return &Client{apiClient}, nil
}
