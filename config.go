package main

import (
	"errors"
	"fmt"
	"os"

	"encoding/json"

	"gopkg.in/yaml.v3"

	"github.com/99designs/keyring"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

type Config struct {
	Name    string `json:"name";yaml:"name"`
	Client  string `json:"client";yaml:"client"`
	Tenant  string `json:"tenant";yaml:"tenant"`
	BaseUrl string `json:"baseUrl";yaml:"baseUrl"`
}

func (config Config) Replace(unmarshaler cache.Unmarshaler, key string) {
	ring, err := keyring.Open(keyring.Config{ServiceName: "msaler"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open keyring: %v\n", err)
	}

	value, err := ring.Get(config.Tenant + config.Client)
	if err != nil {
		if !errors.Is(err, keyring.ErrKeyNotFound) {
			fmt.Fprintf(os.Stderr, "Failed to read keyring: %v\n", err)
		}
	} else {
		if err = unmarshaler.Unmarshal(value.Data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to unmarshal cache: %v\n", err)
		}
	}
}

func (config Config) Export(marshaler cache.Marshaler, key string) {

	bytes, err := marshaler.Marshal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal cache: %v\n", err)
	} else {
		ring, err := keyring.Open(keyring.Config{ServiceName: "msaler"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open keyring: %v\n", err)
		}

		if err := ring.Set(keyring.Item{
			Key:  config.Tenant + config.Client,
			Data: bytes,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write keyring: %v\n", err)
		}
	}
}

func (config *Config) Json() {
	p, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("%s\n", p)
	}
}

func (config *Config) Yaml() {
	p, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("%s\n", p)
	}
}
