package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func listOrganizations() {
	orgs, err := client.ListOrganizations()
	if err != nil {
		logger.Fatal(err)
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "name\tid\tslug")
	for _, org := range orgs {
		fmt.Fprintf(w, "%s\t%s\t%s\n", org.Name, org.Id, org.Slug)
	}
	w.Flush()
}
