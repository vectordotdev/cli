package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/aybabtme/rgbterm"
	"github.com/timberio/cli/api"
)

// Levels as defined by https://github.com/timberio/log-event-json-schema/blob/master/schema.json
type Level string

const (
	Emergency Level = "emergency"
	Alert           = "alert"
	Critical        = "critical"
	Error           = "error"
	Warning         = "warn"
	Notice          = "notice"
	Info            = "info"
	Debug           = "debug"
)

// shortened to 4 characters to not take as much width (given warning is 7)
func (l Level) ShortName() string {
	switch l {
	case Emergency:
		return "emrg"
	case Alert:
		return "alrt"
	case Critical:
		return "crit"
	case Error:
		return "err"
	case Warning:
		return "warn"
	case Notice:
		return "noti"
	case Info:
		return "info"
	case Debug:
		return "debg"
	default:
		return "unkn"
	}
}

// Borrowed from default web UI theme
func (l Level) Color() [3]uint8 {
	switch l {
	case Emergency:
		return [3]uint8{155, 59, 199}
	case Alert:
		return [3]uint8{239, 86, 87}
	case Critical:
		return [3]uint8{239, 86, 87}
	case Error:
		return [3]uint8{239, 86, 87}
	case Warning:
		return [3]uint8{253, 183, 24}
	case Notice:
		return [3]uint8{76, 196, 168}
	case Info:
		return [3]uint8{76, 196, 168}
	case Debug:
		return [3]uint8{66, 126, 219}
	default:
		return [3]uint8{97, 97, 130}
	}
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
func tail(w io.Writer, appIds []string, query string, format string, facets []string, colorize bool) {
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
			if err := formatLine(w, line, fields, facets, colorScale, colorize); err != nil {
				logger.Fatal(err)
			}
		}

		if len(logLines) != 0 {
			datetimeGreaterThan = logLines[len(logLines)-1].Datetime
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func stringContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// TODO this is taking a lot of arguments...
func formatLine(w io.Writer, line *api.LogLine, fields []string, facets []string, colorScale *OrdinalColorScale, colorize bool) error {
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
			level := Level(line.Level)
			formattedField = fmt.Sprintf("%-4s", level.ShortName())
			if colorize {
				levelColor := level.Color()
				formattedField = rgbterm.FgString(formattedField, levelColor[0], levelColor[1], levelColor[2])
			}
		case "message":
			formattedField = line.Message
		default:
			formattedField = findField(strings.Split(field, "."), line.Fields)
			if stringContains(facets, field) && colorize {
				color := colorScale.Get(formattedField)
				formattedField = rgbterm.FgString(formattedField, color[0], color[1], color[2])
			}
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
		return fmt.Sprintf("%v", v)
	}

	fields, ok = v.(map[string]interface{})
	if !ok {
		return ""
	}

	return findField(path[1:], fields)
}
