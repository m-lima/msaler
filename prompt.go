package main

import (
	"errors"
	"os"

	"github.com/manifoldco/promptui"
)

func PromptInput(label string, validator promptui.ValidateFunc, hidden bool) (string, error) {
	var mask rune
	if hidden {
		mask = '*'
	}

	prompt := promptui.Prompt{
		Label:    label,
		Validate: validator,
		Mask:     mask,
		Stdout:   os.Stderr,
	}

	return prompt.Run()
}

func PromptYesNo(label string) (bool, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     []string{"Yes", "No"},
		HideHelp:  true,
		IsVimMode: true,
		Stdout:    os.Stderr,
	}
	value, _, err := prompt.Run()
	return value == 0, err
}

func PromptSelectClient(clients map[string]Client) (string, error) {
	type Item struct {
		Name        string
		Project     string
		BaseUrl     string
		Tenant      string
		Interactive string
	}

	if len(clients) == 0 {
		return "", errors.New("No configured clients")
	}

	i := 0
	list := make([]Item, len(clients))
	for name, client := range clients {
		project := client.Project
		baseUrl := client.BaseUrl
		if len(baseUrl) > 0 {
			baseUrl = " " + baseUrl
		}
		tenant := client.Tenant.Name
		if len(tenant) > 0 {
			tenant = " " + tenant
		}
		interactive := ""
		if !client.WithSecret {
			interactive = " interactive"
		}
		list[i] = Item{
			Name:        name,
			Project:     project,
			BaseUrl:     baseUrl,
			Tenant:      tenant,
			Interactive: interactive,
		}
		i++
	}

	templates := &promptui.SelectTemplates{
		Label:    `{{ "Client:" | blue }}`,
		Active:   "▸ {{ .Name | bold }}",
		Inactive: "  {{ .Name }}",
		Selected: "Client: {{ .Name | green }}",
		Details:  "{{ .Project | cyan }}{{ .BaseUrl }}{{ .Tenant | faint }}{{ .Interactive | bold }}",
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
