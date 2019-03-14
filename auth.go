package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"

	"github.com/mitchellh/go-homedir"
	"github.com/timberio/cli/api"
)

type Credential struct {
	Active           bool
	OrganizationID   string
	OrganizationName string
	APIKey           string
}

var credentialsFileName = "credentials"

// Main function for authenticating and persisting the API key
func auth(apiKey string) (*api.Organization, error) {
	if apiKey == "" {
		return nil, errors.New("API key cannot be blank")
	}

	client = api.NewClient(host, apiKey)

	// Grab the organization for the API key. This is required
	// to build the credentials file and it also validates the
	// API key.
	organization, err := getCurrentOrganization(client)

	// Load the current credentials
	credentials, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	credentialSet := false

	for _, credential := range credentials {
		credential.Active = false
		if credential.OrganizationID == organization.ID {
			credential.OrganizationName = organization.Name
			credential.APIKey = apiKey
			credential.Active = true
			credentialSet = true
		}
	}

	// Add the new credential
	if !credentialSet {
		credential := &Credential{
			Active:           true,
			OrganizationID:   organization.ID,
			OrganizationName: organization.Name,
			APIKey:           apiKey,
		}

		credentials = append(credentials, credential)
	}

	err = saveCredentials(credentials)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

// Lists all credentials stored on the user's machine
func listCredentials() error {
	credentials, err := loadCredentials()
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "Active\tOrg ID\tOrg Name\tAPI Key")
	for _, credential := range credentials {
		activeMarker := ""
		if credential.Active {
			activeMarker = "  *  "
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", activeMarker, credential.OrganizationID, credential.OrganizationName, credential.APIKey[0:8]+"...")
	}
	w.Flush()

	return nil
}

func deleteCredential(orgID string) error {
	credentials, err := loadCredentials()
	if err != nil {
		return err
	}

	i := 0 // output index
	for _, credential := range credentials {
		if credential.OrganizationID != orgID {
			credentials[i] = credential
			i++
		}
	}

	credentials = credentials[:i]

	return saveCredentials(credentials)
}

func switchActiveCredentials(orgID string) error {
	credentials, err := loadCredentials()
	if err != nil {
		return err
	}

	for _, credential := range credentials {
		if credential.OrganizationID == orgID {
			credential.Active = true
		}
	}

	err = saveCredentials(credentials)
	if err != nil {
		return err
	}

	successWriter.Write([]byte("Active credential successfully switched"))

	return nil
}

// Get the active credential. If no credential is marked
// as active the first credential will be returned. And if there
// are not credentials `nil` will be returned.
func getActiveCredential() (*Credential, error) {
	credentials, err := loadCredentials()
	if err != nil {
		return nil, err
	}

	for _, credential := range credentials {
		if credential.Active {
			return credential, nil
		}
	}

	if len(credentials) == 0 {
		return nil, nil
	} else {
		return credentials[0], nil
	}
}

//
// Util
//

func getCurrentOrganization(client *api.Client) (*api.Organization, error) {
	// Grab the organization for the API key. This is required
	// to build the credentials file and it also validates the
	// API key.
	organizations, err := client.ListOrganizations()
	if err != nil {
		return nil, err
	}

	organization := organizations[0]
	return organization, nil
}

func getTimberDirPath() (string, error) {
	// Grab the home directory since we'll be installing timber
	// state files there.
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	// Ensure the timber directory exists
	timberDir := path.Join(homeDir, ".timber")
	return timberDir, nil
}

func saveCredentials(credentials []*Credential) error {
	// Ensure the timber directory exists
	timberDir, err := getTimberDirPath()
	err = os.MkdirAll(timberDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Ensure that the credentials file exists
	credentialsPath := path.Join(timberDir, credentialsFileName)
	f, err := os.Create(credentialsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Save the credentials
	json, err := json.MarshalIndent(credentials, "", "	")
	if err != nil {
		return err
	}

	_, err = f.Write(json)
	if err != nil {
		return err
	}

	return nil
}

func loadCredentials() ([]*Credential, error) {
	var credentials []*Credential

	timberDirPath, err := getTimberDirPath()
	if err != nil {
		return nil, err
	}

	credentialsPath := path.Join(timberDirPath, credentialsFileName)

	credentialsJson, err := ioutil.ReadFile(credentialsPath)
	if os.IsNotExist(err) {
		return credentials, nil
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(credentialsJson, &credentials)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}
