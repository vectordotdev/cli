package main

import "time"

type tailRequest struct {
	ApplicationIds      []string  `json:"application_ids"`
	DatetimeGreaterThan time.Time `json:"dt_gt"`
	Limit               int       `json:"limit"`
	Sort                string    `json:"sort"` // TODO maybe make this an "enum"
}

type tailResponse struct {
	LogLines            []*logLine `json:"data"`
	QueryStringElements []string   `json:"query_string_elements"`
}

type logLine struct {
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
