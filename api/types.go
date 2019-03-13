package api

import "fmt"
import "time"

// TODO maybe handle nullable fields better

type ServiceError struct {
	StatusCode int
	Body       string
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("Request to Timber API failed! Status: %d, Body: %s", e.StatusCode, e.Body)
}

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
	OrganizationId        string    `json:"organization_id"`
	PlatformType          string    `json:"platform_type"`
	Slug                  string    `json:"slug"`
	SourceType            string    `json:"source_type"`
	Tags                  []string  `json:"tags"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type Organization struct {
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

type LogLine struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"application_id"`
	Datetime      time.Time `json:"dt"`
	Level         string    `json:"level"`
	Message       string    `json:"message"`

	Fields map[string]interface{}
}

// TODO fill out chart settings
type SavedView struct {
	ChartSettings   interface{} `json:"chart_settings"`
	ConsoleSettings struct {
		DtGte         *string  `json:"dt_gte"`
		DtLte         *string  `json:"dt_lte"`
		Facets        []string `json:"facets"`
		LogLineFormat string   `json:"log_line_format"`
		Query         *string  `json:"query"`
		SourceIds     []string `json:"source_ids"`
	} `json:"console_settings"`
	Id             string `json:"id"`
	Name           string `json:"name"`
	OrganizationId string `json:"organization_id"`
	Type           string `json:"type"`
}
