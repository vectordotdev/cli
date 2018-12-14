package main

import (
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

func tail(appIds []string) {
	datetimeGreaterThan := time.Now().Add(-5 * time.Minute) // TODO make a flag?
	for {
		logLines, err := client.Search(appIds, datetimeGreaterThan)
		if err != nil {
			logger.Fatal(err)
		}

		// Example:
		// Dec 14 09:50:16am info ec2-54-175-235-51 Frame batch read, size: 41, iterator_age_ms: 0
		for _, line := range logLines {
			fmt.Printf("%s %s %s %s\n", line.Datetime.Format("Jan 02 03:04:05pm"), severityToLevel(line.Severity), line.Context.System.Hostname, line.Message)
		}

		if len(logLines) != 0 {
			datetimeGreaterThan = logLines[len(logLines)-1].Datetime
		}

		time.Sleep(2 * time.Second)
	}
}
