package api

import "time"

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
	Severity      int       `json:"severity"` // TODO remove and depend on Level
	Message       string    `json:"message"`

	Fields map[string]interface{}
}
