package config

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	service = "paymostats"
	account = "apiKey"
)

var ErrNoApiKey = errors.New("no api key configured")

func ResolveApiKey() (string, error) {
	// env (mainly useful for dev)
	if v := os.Getenv("PAYMOSTATS_API_KEY"); v != "" {
		return v, nil
	}
	// Keychain
	v, err := keyring.Get(service, account)
	if err == keyring.ErrNotFound {
		return "", ErrNoApiKey
	}
	return v, err
}

func SaveApiKey(tok string) error {
	return keyring.Set(service, account, tok)
}

func DeleteApiKey() error {
	return keyring.Delete(service, account)
}
