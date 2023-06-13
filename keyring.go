package main

import (
	"github.com/99designs/keyring"
)

type Keyring struct {
	ring keyring.Keyring
	key  string
}

func OpenKeyring(clientId string) (Keyring, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "msaler",
	})

	if err != nil {
		return Keyring{}, err
	} else {
		key := "client_secret_" + clientId
		return Keyring{ring, key}, nil
	}
}

func (ring Keyring) Save(secret string) error {
	item := keyring.Item{
		Key:  ring.key,
		Data: []byte(secret),
	}
	return ring.ring.Set(item)
}

func (ring Keyring) Load() (string, error) {
	secretEntry, err := ring.ring.Get(ring.key)
	if err != nil {
		return "", err
	} else {
		return string(secretEntry.Data), nil
	}
}

func (ring Keyring) Remove() error {
	return ring.ring.Remove(ring.key)
}
