package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/aybabtme/rgbterm"
	"github.com/timberio/timber-cli/api"
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

var tokenRegexp = regexp.MustCompile(`{{\s*(.*?)\s*}}`)

// TODO fallback to 16 colors
// TODO implement format parser
//	Currently only supports a format made of identifiers, space delimited
func tail(w io.Writer, appIds []string, query string, format string, colorize bool) {
	fields := []string{}
	for _, match := range tokenRegexp.FindAllStringSubmatch(format, -1) {
		fields = append(fields, match[1])
	}

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
			if err := formatLine(w, line, fields, colorScale, colorize); err != nil {
				logger.Fatal(err)
			}
		}

		if len(logLines) != 0 {
			datetimeGreaterThan = logLines[len(logLines)-1].Datetime
		}

		time.Sleep(2 * time.Second)
	}
}

//fmt.Fprintf(w, "%s %s %s %s\n",
//datetime,
//level,
//hostname,
//line.Message,
//)

// TODO this is taking a lot of arguments...
func formatLine(w io.Writer, line *api.LogLine, fields []string, colorScale *OrdinalColorScale, colorize bool) error {
	for _, field := range fields {
		formattedField := ""
		switch field {
		case "date":
			formattedField = line.Datetime.Format("Jan 02 03:04:05pm")
			if colorize {
				formattedField = rgbterm.FgString(formattedField, 85, 79, 201)
			}
		case "context.system.hostname":
			formattedField = fmt.Sprintf("%-20s", findField(strings.Split(field, "."), line.Fields))
			if colorize {
				hostnameColor := colorScale.Get(formattedField)
				formattedField = rgbterm.FgString(formattedField, hostnameColor[0], hostnameColor[1], hostnameColor[2])
			}
		case "level":
			//TODO replace with level
			severity := Severity(line.Severity)
			formattedField = fmt.Sprintf("%-4s", severity.Name())
			if colorize {
				severityColor := severity.Color()
				formattedField = rgbterm.FgString(formattedField, severityColor[0], severityColor[1], severityColor[2])
			}
		case "message":
			formattedField = line.Message
		default:
			formattedField = findField(strings.Split(field, "."), line.Fields)
		}

		fmt.Fprintf(w, "%s ", formattedField)
	}

	fmt.Fprintln(w)

	return nil
}

// given a path in the form of []string{"path", "to", "value"}, extract this value from fields
// if the value cannot be found at the path, returns ""
func findField(path []string, fields map[string]interface{}) string {
	if len(path) == 0 {
		return ""
	}

	v, ok := fields[path[0]]
	if !ok {
		return ""
	}

	if len(path) == 1 {
		return fmt.Sprintf("%s", v)
	}

	fields, ok = v.(map[string]interface{})
	if !ok {
		return ""
	}

	return findField(path[1:], fields)
}
