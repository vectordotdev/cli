package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// Severities defined by Syslog 5424 as strings
// Index corresponds to numerical severity
var severities = []string{
	"emerg",
	"alert",
	"crit",
	"err",
	"warning",
	"notice",
	"info",
	"debug",
}

func severityToLevel(i int) string {
	if i < 0 || i > len(severities) {
		return "unknown"
	}

	return severities[i]
}

type tailRequest struct {
	ApplicationIds      []int     `json:"application_ids"`
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

func tail(host string, apiKey string, appIds []int) {
	datetimeGreaterThan := time.Now().Add(-5 * time.Minute) // TODO make a flag?
	for {
		url := fmt.Sprintf("%s%s", host, "/log_lines/search")
		limit := 250

		r := tailRequest{
			ApplicationIds:      appIds,
			DatetimeGreaterThan: datetimeGreaterThan,
			Limit:               limit,
			Sort:                "dt.desc",
		}

		rBody, err := json.Marshal(r)
		if err != nil {
			logger.Fatal(err)
		}

		resp, err := request("POST", url, rBody, apiKey)
		if err != nil {
			logger.Fatal(err)
		}

		response := tailResponse{
			LogLines: make([]*logLine, limit),
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			logger.Fatal(err)
		}
		_ = resp.Body.Close()

		// Example:
		// Dec 14 09:50:16am info ec2-54-175-235-51 Frame batch read, size: 41, iterator_age_ms: 0
		for _, line := range response.LogLines {
			fmt.Printf("%s %s %s %s\n", line.Datetime.Format("Jan 02 03:04:05pm"), severityToLevel(line.Severity), line.Context.System.Hostname, line.Message)
		}

		if len(response.LogLines) != 0 {
			datetimeGreaterThan = response.LogLines[len(response.LogLines)-1].Datetime
		}

		time.Sleep(2 * time.Second)
	}
}
