package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"gopkg.in/yaml.v3"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

var ctx = context.Background()

func GetClientToken(args []string) error {
	verbose := false
	for i, arg := range args {
		if arg == "-v" {
			verbose = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) > 1 {
		return errors.New("Too many arguments")
	}

	clients, err := LoadClients()
	if err != nil {
		return err
	}

	var clientName string
	if len(args) == 0 {
		if clientName, err = PromptSelectClient(clients); err != nil {
			return err
		}
	} else {
		clientName = args[0]
	}

	client, ok := clients[clientName]
	if !ok {
		clientNames := ""
		for name := range clients {
			clientNames += "\n" + name
		}
		return fmt.Errorf("Client name `%s` was not found\nPossible values:\n%s", clientName, clientNames)
	}

	auth, err := getToken(client)
	if err != nil {
		return err
	}

	if verbose {
		asJson, err := json.MarshalIndent(auth.token, "", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", asJson)
		} else {
			fmt.Fprintf(os.Stderr, "%#v\n", auth)
		}
	}

	fmt.Print(auth.accessToken)

	return nil
}

type AuthToken struct {
	token       any
	accessToken string
}

func NewClient(args []string) error {
	if len(args) != 0 {
		return errors.New("Too many arguments")
	}

	clients, _ := LoadClients()

	uuidValidator := func(input string) error {
		if getUuidMatcher().MatchString(input) {
			return nil
		}
		return errors.New("Invalid UUID")
	}

	urlValidator := func(input string) error {
		if getUrlMatcher().MatchString(input) {
			return nil
		}
		return errors.New("Invalid URL")
	}

	nameValidator := func(input string) error {
		if _, ok := clients[input]; ok {
			return errors.New("Name already exists")
		}
		return nil
	}

	name, err := PromptInput("Name", nameValidator, false)
	if err != nil {
		return err
	}

	client, err := PromptInput("Client ID", uuidValidator, false)
	if err != nil {
		return err
	}
	client = strings.ToLower(client)

	tenantId, err := PromptInput("Tenant ID", uuidValidator, false)
	if err != nil {
		return err
	}

	tenantName, err := PromptInput("Tenant Name", nil, false)
	if err != nil {
		return err
	}

	project, err := PromptInput("Project", nil, false)
	if err != nil {
		return err
	}

	baseUrl, err := PromptInput("Base URL", urlValidator, false)
	if err != nil {
		return err
	}

	withSecret, err := PromptYesNo("Use Client Secret")
	if err != nil {
		return err
	}

	if withSecret {
		ring, err := OpenKeyring(client)
		if err == nil {
			secret, err := PromptInput("Client Secret", nil, true)
			if err == nil {
				err = ring.Save(secret)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	clients[name] =
		Client{
			Id:      strings.ToLower(client),
			Project: project,
			Tenant: Tenant{
				Id:   strings.ToLower(tenantId),
				Name: tenantName,
			},
			BaseUrl:    baseUrl,
			WithSecret: withSecret,
		}

	return SaveClients(clients)
}

func DeleteClient(args []string) error {
	if len(args) > 1 {
		return errors.New("Too many arguments")
	}

	clients, err := LoadClients()
	if err != nil {
		return err
	}

	var clientName string
	if len(args) == 0 {
		if clientName, err = PromptSelectClient(clients); err != nil {
			return err
		}
	} else {
		clientName = args[0]
	}

	if _, ok := clients[clientName]; ok {
		delete(clients, clientName)
		SaveClients(clients)
	} else {
		clientNames := ""
		for name := range clients {
			clientNames += "\n" + name
		}
		return fmt.Errorf("Client name `%s` was not found\nPossible values:\n%s", clientName, clientNames)
	}

	return nil
}

func UncacheClient(args []string) error {
	if len(args) > 1 {
		return errors.New("Too many arguments")
	}

	clients, err := LoadClients()
	if err != nil {
		return err
	}

	var clientName string
	if len(args) == 0 {
		if clientName, err = PromptSelectClient(clients); err != nil {
			return err
		}
	} else {
		clientName = args[0]
	}

	client, ok := clients[clientName]
	if !ok {
		clientNames := ""
		for name := range clients {
			clientNames += "\n" + name
		}
		return fmt.Errorf("Client name `%s` was not found\nPossible values:\n%s", clientName, clientNames)
	}

	if client.WithSecret {
		ring, err := OpenKeyring(client.Id)
		if err == nil {
			err = ring.Remove()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	msalClient, err := public.New(client.Id, public.WithAuthority("https://login.microsoftonline.com/"+client.Tenant.Id), public.WithCache(client))

	accounts, err := msalClient.Accounts(ctx)
	if err != nil {
		return err
	}

	if len(accounts) > 0 {
		msalClient.RemoveAccount(ctx, accounts[0])
	}

	return nil
}

func PrintClient(args []string) error {
	if len(args) > 1 {
		return errors.New("Too many arguments")
	}

	clients, err := LoadClients()
	if err != nil {
		return err
	}

	var clientName string
	if len(args) == 0 {
		if clientName, err = PromptSelectClient(clients); err != nil {
			return err
		}
	} else {
		clientName = args[0]
	}

	client, ok := clients[clientName]
	if !ok {
		clientNames := ""
		for name := range clients {
			clientNames += "\n" + name
		}
		return fmt.Errorf("Client name `%s` was not found\nPossible values:\n%s", clientName, clientNames)
	}

	bytes, err := yaml.Marshal(client)
	if err == nil {
		fmt.Print(string(bytes))
	} else {
		fmt.Printf("%+v\n", client)
	}

	return nil
}

func getToken(client Client) (AuthToken, error) {
	if client.WithSecret {
		token, err := getTokenOauth(client)
		if err != nil {
			return AuthToken{}, err
		} else {
			return AuthToken{
					token:       token,
					accessToken: token.AccessToken,
				},
				nil
		}
	} else {
		token, err := getTokenInteractive(client)
		if err != nil {
			return AuthToken{}, err
		} else {
			return AuthToken{
					token:       token,
					accessToken: token.AccessToken,
				},
				nil
		}
	}
}

func getTokenInteractive(client Client) (public.AuthResult, error) {
	msalClient, err := public.New(client.Id, public.WithAuthority("https://login.microsoftonline.com/"+client.Tenant.Id), public.WithCache(client))
	if err != nil {
		return public.AuthResult{}, err
	}

	scopes := []string{client.BaseUrl + ".default"}

	accounts, err := msalClient.Accounts(ctx)
	if err != nil {
		return public.AuthResult{}, err
	}

	if len(accounts) > 0 {
		auth, err := msalClient.AcquireTokenSilent(ctx, scopes, public.WithSilentAccount(accounts[0]))
		if err != nil {
			return msalClient.AcquireTokenInteractive(ctx, scopes)
		}
		return auth, nil
	} else {
		return msalClient.AcquireTokenInteractive(ctx, scopes)
	}
}

func getTokenOauth(client Client) (*oauth2.Token, error) {
	ring, err := OpenKeyring(client.Id)
	if err != nil {
		return nil, err
	}

	secret, err := ring.Load()
	if err != nil || secret == "" {
		secret, err = PromptInput("Client Secret", nil, true)
		if err != nil {
			return nil, err
		}

		if err := ring.Save(secret); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	scopes := []string{client.BaseUrl + ".default"}

	config := clientcredentials.Config{
		ClientID:     client.Id,
		Scopes:       scopes,
		ClientSecret: secret,
		TokenURL:     "https://login.microsoftonline.com/" + client.Tenant.Id + "/oauth2/v2.0/token",
	}

	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	return token, nil
}
