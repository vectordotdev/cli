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
	OrganizationID        string    `json:"organization_id"`
	PlatformType          string    `json:"platform_type"`
	Slug                  string    `json:"slug"`
	SourceType            string    `json:"source_type"`
	Tags                  []string  `json:"tags"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type LogLine struct {
	ApplicationID string `json:"application_id"`
	Context       struct {
		Custom  map[string]interface{} `json:"custom"`
		Runtime struct {
			Application string `json:"application"`
			File        string `json:"file"`
			Function    string `json:"function"`
			Line        int    `json:"line"`
			ModuleName  string `json:"module_name"`
			VMPid       string `json:"vm_pid"`
		} `json:"runtime"`
		System struct {
			Hostname string `json:"hostname"`
			Pid      int    `json:"pid"`
		} `json:"system"`
	} `json:"context"`
	Datetime time.Time `json:"dt"`
	Event    struct {
		Type   string                 `json:"type"`
		Custom map[string]interface{} `json:"custom"`
	} `json:"event"`
	ID       string `json:"id"`
	Level    string `json:"level"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
}
