package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"encoding/json"

	"github.com/99designs/keyring"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/cache"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

const MsalCache = "/tmp/msal_cache"

var (
	KnownConfigs = map[string]Config{
		"jupyter": {
			client:  "a5f77a6e-73a5-4fed-8fc3-818e4d929020",
			tenant:  "65db1639-116f-48c3-9a8b-59f0c228f263",
			baseUrl: "https://greenfield.cognitedata.com/",
		},
		"celo": {
			client:  "1c65d0ae-06a1-4a2e-9dce-81a5362e6972",
			tenant:  "65db1639-116f-48c3-9a8b-59f0c228f263",
			baseUrl: "https://greenfield.cognitedata.com/",
		},
		"demo": {
			client:  "62d51730-37d6-430c-b3c5-d2bcaaf4bdb1",
			tenant:  "d144e8ad-92a5-49c7-9e33-02e965f9679e",
			baseUrl: "https://greenfield.cognitedata.com/",
		},
	}
)

type Config struct {
	client  string
	tenant  string
	baseUrl string
}

func (config Config) Replace(unmarshaler cache.Unmarshaler, key string) {
	ring, err := keyring.Open(keyring.Config{ServiceName: "msaler"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open keyring: %v\n", err)
	}

	value, err := ring.Get(config.tenant + config.client)
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
			Key:  config.tenant + config.client,
			Data: bytes,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write keyring: %v\n", err)
		}
	}
}

func getToken(config Config) (public.AuthResult, error) {
	client, err := public.New(config.client, public.WithAuthority("https://login.microsoftonline.com/"+config.tenant), public.WithCache(config))
	if err != nil {
		return public.AuthResult{}, err
	}

	scopes := []string{config.baseUrl + ".default"}

	accounts := client.Accounts()
	if len(accounts) > 0 {
		return client.AcquireTokenSilent(context.Background(), scopes, public.WithSilentAccount(accounts[0]))
	} else {
		return client.AcquireTokenInteractive(context.Background(), scopes)
	}
}

func getConfig(args []string) (Config, error) {
	if len(args) == 1 {
		config, ok := KnownConfigs[args[0]]
		if !ok {
			return Config{}, errors.New("Configuration not known")
		}
		return config, nil
	} else if len(args) == 3 {
		return Config{
			client:  args[0],
			tenant:  args[1],
			baseUrl: args[2],
		}, nil
	} else {
		return Config{}, errors.New("Wrong number of arguments\nExpected: (well-known | (clientId tenantId baseUrl))")
	}
}

func main() {

	args := os.Args[1:]

	verbose := false
	for i, arg := range args {
		if arg == "-v" {
			verbose = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	config, err := getConfig(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}

	auth, err := getToken(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get token: %v\n", err)
		os.Exit(-1)
	}

	if verbose {
		asJson, err := json.MarshalIndent(auth, "", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", asJson)
		} else {
			fmt.Fprintf(os.Stderr, "%#v\n", auth)
		}
	}

	fmt.Println(auth.AccessToken)
}
