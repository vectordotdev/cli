package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func listSavedViews() error {
	savedViews, err := client.ListSavedViews()
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "name\tid\tsource ids\tfacets\tquery\tformat")
	for _, savedView := range savedViews {
		query := ""
		if savedView.ConsoleSettings.Query != nil {
			query = *savedView.ConsoleSettings.Query
		}
		fmt.Fprintln(w, strings.Join([]string{
			savedView.Name,
			savedView.ID,
			strings.Join(savedView.ConsoleSettings.SourceIds, ","),
			strings.Join(savedView.ConsoleSettings.Facets, ","),
			query,
			fmt.Sprintf(`"%s"`, savedView.ConsoleSettings.LogLineFormat),
		}, "\t"))
	}
	w.Flush()

	return nil
}
