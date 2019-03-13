package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func listSources() error {
	applications, err := client.ListSources()
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "name\tid\tslug\tenvironment\tsource type")
	for _, application := range applications {
		fmt.Fprintln(w, strings.Join([]string{
			application.Name,
			application.Id,
			application.Slug,
			application.Environment,
			application.SourceType,
		}, "\t"))
	}
	w.Flush()

	return nil
}
