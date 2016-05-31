package letsencrypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	lego "github.com/xenolf/lego/acme"
)

func generatePrivateKey(keyType lego.KeyType, file string) (crypto.PrivateKey, error) {
	var privateKey crypto.PrivateKey
	var err error

	switch keyType {
	case lego.EC256:
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case lego.EC384:
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case lego.RSA2048:
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	case lego.RSA4096:
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	case lego.RSA8192:
		privateKey, err = rsa.GenerateKey(rand.Reader, 8192)
	default:
		return nil, fmt.Errorf("Invalid KeyType: %s", keyType)
	}

	if err != nil {
		return nil, err
	}

	var pemBlock *pem.Block

	switch key := privateKey.(type) {
	case *ecdsa.PrivateKey:
		keyBytes, _ := x509.MarshalECPrivateKey(key)
		pemBlock = &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
	case *rsa.PrivateKey:
		pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, err
	}

	pem.Encode(certOut, pemBlock)
	certOut.Close()

	return privateKey, nil
}

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, fmt.Errorf("Unknown private key type.")
}

func getPEMCertSerialNumber(cert []byte) (string, error) {
	pemBlock, _ := pem.Decode(cert)
	if pemBlock == nil {
		return "", fmt.Errorf("Pem decode did not yield a valid block")
	}

	pCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return "", err
	}

	return pCert.SerialNumber.String(), nil
}
