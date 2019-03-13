package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/timberio/cli/api"
)

func setApiKey(apiKey string) (string, error) {
	if apiKey == "" {
		return "", errors.New("API key cannot be blank")
	}

	client = api.NewClient(host, apiKey)

	_, err := client.ListOrganizations()
	if err != nil {
		return "", err
	}

	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	credentialsPath := fmt.Sprint(homeDir, "/.timber")

	f, err := os.Create(credentialsPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents := []byte(apiKey)
	_, err = f.Write(contents)
	if err != nil {
		return "", err
	}

	return credentialsPath, nil
}

func fetchAPIKey() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	credentialsPath := fmt.Sprint(homeDir, "/.timber")

	b, err := ioutil.ReadFile(credentialsPath)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return string(b), nil
}
