package main

import (
	"fmt"
	"io"
	"time"

	"github.com/aybabtme/rgbterm"
)

// Severities defined by Syslog 5424
type Severity int

const (
	Emerg Severity = iota
	Alert
	Crit
	Err
	Warning
	Notice
	Info
	Debug
)

// shortened to 4 characters to not take as much width (given warning is 7)
var severityNames = []string{
	"emrg",
	"alrt",
	"crit",
	"err",
	"warn",
	"noti",
	"info",
	"debg",
}

func (s Severity) Name() string {
	if int(s) < 0 || int(s) > len(severityNames) {
		return "unknown"
	}

	return severityNames[s]
}

// Index corresponds to numerical severity
// Borrowed from default web UI theme
var severityColors = [][3]uint8{
	{155, 59, 199},
	{239, 86, 87},
	{239, 86, 87},
	{239, 86, 87},
	{253, 183, 24},
	{76, 196, 168},
	{76, 196, 168},
	{66, 126, 219},
}

func (s Severity) Color() [3]uint8 {
	if int(s) < 0 || int(s) > len(severityNames) {
		return [3]uint8{97, 97, 130}
	}

	return severityColors[s]
}

// Used for coloring distinct values that are not apriori known
// Borrowed from default web UI theme
var ordinalScale = [][3]uint8{
	{218, 144, 62},
	{225, 110, 111},
	{90, 189, 152},
	{80, 155, 251},
	{158, 83, 221},
}

// TODO handle outputing without colors
// TODO: fallback to 16 colors
func tail(w io.Writer, appIds []string, query string, colorize bool) {
	colorScale := NewOrdinalColorScale(ordinalScale)
	datetimeGreaterThan := time.Now().Add(-5 * time.Minute) // TODO make a flag?
	for {
		logLines, err := client.Search(appIds, datetimeGreaterThan, query)
		if err != nil {
			logger.Fatal(err)
		}

		// Example:
		// Dec 14 09:50:16am info ec2-54-175-235-51 Frame batch read, size: 41, iterator_age_ms: 0
		for _, line := range logLines {
			hostname := fmt.Sprintf("%-20s", line.Context.System.Hostname)
			if colorize {
				hostnameColor := colorScale.Get(line.Context.System.Hostname)
				hostname = rgbterm.FgString(hostname, hostnameColor[0], hostnameColor[1], hostnameColor[2])
			}

			severity := Severity(line.Severity)
			level := fmt.Sprintf("%-4s", severity.Name())
			if colorize {
				severityColor := severity.Color()
				level = rgbterm.FgString(level, severityColor[0], severityColor[1], severityColor[2])
			}

			datetime := line.Datetime.Format("Jan 02 03:04:05pm")
			if colorize {
				datetime = rgbterm.FgString(datetime, 85, 79, 201)
			}

			fmt.Fprintf(w, "%s %s %s %s\n",
				datetime,
				level,
				hostname,
				line.Message,
			)
		}

		if len(logLines) != 0 {
			datetimeGreaterThan = logLines[len(logLines)-1].Datetime
		}

		time.Sleep(2 * time.Second)
	}
}
