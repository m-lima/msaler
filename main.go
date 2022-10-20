package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/manifoldco/promptui"
)

func getToken(client Client) (public.AuthResult, error) {
	msalClient, err := public.New(client.Id, public.WithAuthority("https://login.microsoftonline.com/"+client.Tenant.Id), public.WithCache(client))
	if err != nil {
		return public.AuthResult{}, err
	}

	scopes := []string{client.BaseUrl + ".default"}

	accounts := msalClient.Accounts()
	if len(accounts) > 0 {
		return msalClient.AcquireTokenSilent(context.Background(), scopes, public.WithSilentAccount(accounts[0]))
	} else {
		return msalClient.AcquireTokenInteractive(context.Background(), scopes)
	}
}

func promptSelectClient(clients map[string]Client) (string, error) {
	type Item struct {
		Name    string
		Project string
		BaseUrl string
		Tenant  string
	}

	if len(clients) == 0 {
		return "", errors.New("No configured clients")
	}

	i := 0
	list := make([]Item, len(clients))
	for name, client := range clients {
		project := client.Project
		if len(project) > 0 {
			project = project + " "
		}
		tenant := client.Tenant.Name
		if len(tenant) > 0 {
			tenant = " " + tenant
		}
		list[i] = Item{
			Name:    name,
			Project: project,
			BaseUrl: client.BaseUrl,
			Tenant:  tenant,
		}
		i++
	}

	templates := &promptui.SelectTemplates{
		Label:    `{{ "Client:" | blue }}`,
		Active:   "▸ {{ .Name | bold }}",
		Inactive: "  {{ .Name }}",
		Selected: "Client: {{ .Name | green }}",
		Details:  "{{ .Name | white }} {{ .Project | cyan }}{{ .BaseUrl }}{{ .Tenant | faint }}",
	}

	prompt := promptui.Select{
		Items:     list,
		IsVimMode: true,
		HideHelp:  true,
		Templates: templates,
		Stdout:    os.Stderr,
	}

	i, _, err := prompt.Run()
	return list[i].Name, err
}

func connectClient(args []string) error {
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
		if clientName, err = promptSelectClient(clients); err != nil {
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
		asJson, err := json.MarshalIndent(auth, "", "  ")
		if err == nil {
			fmt.Fprintf(os.Stderr, "%s\n", asJson)
		} else {
			fmt.Fprintf(os.Stderr, "%#v\n", auth)
		}
	}

	fmt.Println(auth.AccessToken)

	return nil
}

func newClient(args []string) error {
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

	prompt := promptui.Prompt{
		Label:    "Name",
		Validate: nameValidator,
		Stdout:   os.Stderr,
	}
	name, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label:    "Client ID",
		Validate: uuidValidator,
		Stdout:   os.Stderr,
	}
	client, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label:  "Project",
		Stdout: os.Stderr,
	}
	project, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label:    "Tenant ID",
		Validate: uuidValidator,
		Stdout:   os.Stderr,
	}
	tenantId, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label:  "Tenant Name",
		Stdout: os.Stderr,
	}
	tenantName, err := prompt.Run()
	if err != nil {
		return err
	}

	prompt = promptui.Prompt{
		Label:    "Base URL",
		Validate: urlValidator,
		Stdout:   os.Stderr,
	}
	baseUrl, err := prompt.Run()
	if err != nil {
		return err
	}

	clients[name] =
		Client{
			Id:      strings.ToLower(client),
			Project: project,
			Tenant: Tenant{
				Id:   strings.ToLower(tenantId),
				Name: tenantName,
			},
			BaseUrl: baseUrl,
		}

	return SaveClients(clients)
}

func deleteClient(args []string) error {
	if len(args) > 1 {
		return errors.New("Too many arguments")
	}

	clients, err := LoadClients()
	if err != nil {
		return err
	}

	var clientName string
	if len(args) == 0 {
		if clientName, err = promptSelectClient(clients); err != nil {
			return err
		}
	} else {
		clientName = args[1]
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

func main() {
	var err error
	if len(os.Args) == 1 {
		err = fmt.Errorf("Missing command.\nAvailable commands: [token, new, delete]")
	} else {
		modeParam := os.Args[1]

		if "token" == modeParam {
			err = connectClient(os.Args[2:])
		} else if "new" == modeParam {
			newClient(os.Args[2:])
		} else if "delete" == modeParam {
			deleteClient(os.Args[2:])
		} else {
			err = fmt.Errorf("Unrecognized parameters `%s`.\nAvailable commands: [token, new, delete]", strings.Join(os.Args[1:], " "))
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}
