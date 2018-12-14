package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func listApplications() {
	applications, err := client.ListApplications()
	if err != nil {
		logger.Fatal(err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "name\tid\tslug\tenvironment\tplatform type\tsource type")
	for _, application := range applications {
		fmt.Fprintln(w, strings.Join([]string{
			application.Name,
			application.Id,
			application.Slug,
			application.Environment,
			application.PlatformType,
			application.SourceType,
		}, "\t"))
	}
	w.Flush()
}
