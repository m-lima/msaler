package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"encoding/json"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/manifoldco/promptui"
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

func loadKnownConfig(name string) (Config, error) {
	config, ok := KnownConfigs[name]
	if !ok {
		return Config{}, fmt.Errorf("Configuration `%s` not found", name)
	}
	return config, nil
}

func promptCustom() (Config, error) {
	uuidMatcher, _ := regexp.Compile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
	urlMatcher, _ := regexp.Compile("^http[s]?://.+/$")

	uuidValidator := func(input string) error {
		if uuidMatcher.MatchString(input) {
			return nil
		} else {
			return errors.New("Invalid UUID")
		}
	}

	urlValidator := func(input string) error {
		if urlMatcher.MatchString(input) {
			return nil
		} else {
			return errors.New("Invalid URL")
		}
	}

	prompt := promptui.Prompt{
		Label:    "Name",
		Validate: nil,
	}

	prompt = promptui.Prompt{
		Label:    "Client ID",
		Validate: uuidValidator,
	}

	client, err := prompt.Run()
	if err != nil {
		return Config{}, err
	}

	prompt = promptui.Prompt{
		Label:    "Tenant ID",
		Validate: uuidValidator,
	}

	tenant, err := prompt.Run()
	if err != nil {
		return Config{}, err
	}

	prompt = promptui.Prompt{
		Label:    "Base URL",
		Validate: urlValidator,
	}

	baseUrl, err := prompt.Run()
	if err != nil {
		return Config{}, err
	}

	return Config{
		client:  strings.ToLower(client),
		tenant:  strings.ToLower(tenant),
		baseUrl: baseUrl,
	}, nil
}

func getConfig(args []string) (Config, error) {
	if len(args) == 0 {
		items := []string{"celo", "demo", "jupyter", "custom"}
		prompt := promptui.Select{
			Label:     "Config",
			Items:     items,
			IsVimMode: true,
			HideHelp:  true,
		}

		i, config, err := prompt.Run()
		if err != nil {
			return Config{}, err
		}

		if i == len(items)-1 {
			return promptCustom()
		} else {
			return loadKnownConfig(config)
		}
	} else if len(args) == 1 {
		return loadKnownConfig(args[0])
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
