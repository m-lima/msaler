package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/99designs/keyring"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
)

var (
	uuidMatcher *regexp.Regexp
	urlMatcher  *regexp.Regexp
	ring        keyring.Keyring
)

type Tenant struct {
	Id   string `json:"id" yaml:"id"`
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type Client struct {
	Id      string `json:"id" yaml:"id"`
	Project string `json:"project,omitempty" yaml:"project,omitempty"`
	Tenant  Tenant `json:"tenant" yaml:"tenant"`
	BaseUrl string `json:"baseUrl" yaml:"baseUrl"`
}

func getUuidMatcher() *regexp.Regexp {
	if uuidMatcher == nil {
		uuidMatcher, _ = regexp.Compile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
	}
	return uuidMatcher
}

func getUrlMatcher() *regexp.Regexp {
	if urlMatcher == nil {
		urlMatcher, _ = regexp.Compile("^http[s]?://.+/$")
	}
	return urlMatcher
}

func getKeyring() (keyring.Keyring, error) {
	var err error
	if ring == nil {
		ring, err = keyring.Open(keyring.Config{ServiceName: "msaler"})
	}
	return ring, err
}

func (client Client) Replace(unmarshaler cache.Unmarshaler, key string) {
	ring, err := getKeyring()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open keyring: %v\n", err)
		return
	}

	value, err := ring.Get(client.Tenant.Id + client.Id)
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

func (client Client) Export(marshaler cache.Marshaler, key string) {
	bytes, err := marshaler.Marshal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal cache: %v\n", err)
		return
	}

	ring, err := getKeyring()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open keyring: %v\n", err)
		return
	}

	if err := ring.Set(keyring.Item{
		Key:  client.Tenant.Id + client.Id,
		Data: bytes,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write keyring: %v\n", err)
	}
}

func ConfigPath() (string, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configPath, "msaler", "msaler.yaml"), nil
}

func LoadClients() (map[string]Client, error) {
	clients := make(map[string]Client)
	configPath, err := ConfigPath()
	if err != nil {
		return clients, err
	}

	bytes, err := os.ReadFile(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return clients, err
		}
		return clients, nil
	}

	if err = yaml.Unmarshal(bytes, &clients); err != nil {
		return clients, err
	}

	uuidMatcher := getUuidMatcher()
	urlMatcher := getUrlMatcher()

	var error []string
	for k, v := range clients {
		if k == "" {
			error = append(error, "Configuration with empty name")
		}
		if !uuidMatcher.MatchString(v.Id) {
			error = append(error, fmt.Sprintf("Configuration for `%s` has invalid id: %s", k, v.Id))
		}
		if !uuidMatcher.MatchString(v.Tenant.Id) {
			error = append(error, fmt.Sprintf("Configuration for `%s` has invalid tenant id: %s", k, v.Tenant.Id))
		}
		if !urlMatcher.MatchString(v.BaseUrl) {
			error = append(error, fmt.Sprintf("Configuration for `%s` has invalid base URL: %s", k, v.BaseUrl))
		}
	}

	if len(error) > 0 {
		return clients, errors.New(strings.Join(error, "\n"))
	} else {
		return clients, nil
	}
}

func SaveClients(clients map[string]Client) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	bytes, err := yaml.Marshal(clients)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if _, err = os.Stat(configDir); errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir(configDir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(configPath, bytes, 0644)
}
