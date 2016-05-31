package letsencrypt

import (
	"crypto"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
	lego "github.com/xenolf/lego/acme"
)

type Account struct {
	Email        string                     `json:"email"`
	Registration *lego.RegistrationResource `json:"registrations"`

	key  crypto.PrivateKey
	path string
}

// NewAccount creates a new or gets a stored LE account for the given email
func NewAccount(email string, apiVer ApiVersion, keyType lego.KeyType) (*Account, error) {
	accPath := accountPath(email, apiVer)
	keyFile := path.Join(accPath, "account.key")
	accountFile := path.Join(accPath, "account.json")

	var privKey crypto.PrivateKey
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		logrus.Infof("Generating private key (%s) for %s.", keyType, email)
		privKey, err = generatePrivateKey(keyType, keyFile)
		if err != nil {
			return nil, fmt.Errorf("Error generating private key: %v", err)
		}
		logrus.Debugf("Saved account key to %s", keyFile)
	} else {
		privKey, err = loadPrivateKey(keyFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading private key from %s: %v", keyFile, err)
		}
	}

	if _, err := os.Stat(accountFile); os.IsNotExist(err) {
		return &Account{Email: email, key: privKey, path: accPath}, nil
	}

	fileBytes, err := ioutil.ReadFile(accountFile)
	if err != nil {
		return nil, fmt.Errorf("Could not load account config file: %v", err)
	}

	var acc Account
	err = json.Unmarshal(fileBytes, &acc)
	if err != nil {
		return nil, fmt.Errorf("Could not parse account config file: %v", err)
	}

	acc.key = privKey
	acc.path = accPath
	return &acc, nil
}

// Save the account to disk
func (a *Account) Save() error {
	jsonBytes, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		return err
	}
	accountFile := path.Join(a.path, "account.json")
	return ioutil.WriteFile(accountFile, jsonBytes, 0700)
}

/* Methods implementing the lego.User interface*/

// GetEmail returns the email address for the account
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.key
}

// GetRegistration returns the server registration
func (a *Account) GetRegistration() *lego.RegistrationResource {
	return a.Registration
}

func accountPath(email string, apiVer ApiVersion) string {
	path := path.Join(StorageDir, strings.ToLower(string(apiVer)), "accounts", email)
	maybeCreatePath(path)
	return path
}
