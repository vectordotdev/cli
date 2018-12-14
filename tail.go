package main

import (
	"fmt"
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

//TODO: fallback to 16 colors
//Used for coloring distinct values that are not apriori known
var ordinalScale = [][3]uint8{
	{218, 144, 62},
	{225, 110, 111},
	{90, 189, 152},
	{80, 155, 251},
	{158, 83, 221},
}

type OrdinalColorScale struct {
	colors [][3]uint8
	index  map[string]int
	curr   int
}

func NewOrdinalColorScale(colors [][3]uint8) *OrdinalColorScale {
	return &OrdinalColorScale{
		colors: colors,
		index:  map[string]int{},
	}
}

func (o *OrdinalColorScale) Get(s string) [3]uint8 {
	i, ok := o.index[s]
	if !ok {
		i = o.curr
		o.index[s] = i
		o.curr = (o.curr + 1) % len(o.colors)
	}

	return o.colors[i]
}

// TODO handle outputing without colors
func tail(appIds []string, query string) {
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
			hostnameColor := colorScale.Get(line.Context.System.Hostname)
			hostname := rgbterm.FgString(fmt.Sprintf("%-20s", line.Context.System.Hostname), hostnameColor[0], hostnameColor[1], hostnameColor[2])

			severity := Severity(line.Severity)
			severityColor := severity.Color()
			level := rgbterm.FgString(fmt.Sprintf("%-4s", severity.Name()), severityColor[0], severityColor[1], severityColor[2])

			fmt.Printf("%s %s %s %s\n",
				rgbterm.FgString(line.Datetime.Format("Jan 02 03:04:05pm"), 85, 79, 201),
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
