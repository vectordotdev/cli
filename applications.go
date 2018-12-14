package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type applicationResponse struct {
	Applications []*Application `json:"data"`
}

// TODO maybe handle nullable fields better
type Application struct {
	APIKey                string    `json:"api_key"`
	BillingSubscriptionID string    `json:"billing_subscription_id"`
	Environment           string    `json:"environment"`
	ExternalProvider      *string   `json:"external_provider"`
	Id                    string    `json:"id"`
	InsertedAt            time.Time `json:"inserted_at"`
	LanguageType          *string   `json:"language_type"`
	LogLineFormat         string    `json:"log_line_format"`
	Name                  string    `json:"name"`
	OrganizationID        string    `json:"organization_id"`
	PlatformType          string    `json:"platform_type"`
	Slug                  string    `json:"slug"`
	SourceType            string    `json:"source_type"`
	Tags                  []string  `json:"tags"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func getApplications(host string, apiKey string) []*Application {
	url := fmt.Sprintf("%s%s", host, "/applications")
	resp, err := request("GET", url, nil, apiKey)
	if err != nil {
		logger.Fatal(err)
	}

	response := applicationResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Fatal(err)
	}
	_ = resp.Body.Close()

	return response.Applications
}
