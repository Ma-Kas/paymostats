package config

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	service = "paymostats"
	account = "token"
)

var ErrNoToken = errors.New("no token configured")

func ResolveToken() (string, error) {
	// env (mainly useful for dev)
	if v := os.Getenv("PAYMOSTATS_TOKEN"); v != "" {
		return v, nil
	}
	// Keychain
	v, err := keyring.Get(service, account)
	if err == keyring.ErrNotFound {
		return "", ErrNoToken
	}
	return v, err
}

func SaveToken(tok string) error {
	return keyring.Set(service, account, tok)
}

func DeleteToken() error {
	return keyring.Delete(service, account)
}
