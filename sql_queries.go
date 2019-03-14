package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/timberio/cli/api"
	"github.com/tj/go-spin"
)

func executeSQLQuery(query string) error {
	organization, err := getCurrentOrganization(client)
	if err != nil {
		return err
	}

	sqlQuery, err := client.CreateSQLQuery(organization.ID, query)
	if err != nil {
		return err
	}

	sqlQuery, err = waitForSQLQuery(sqlQuery)
	if err != nil {
		return err
	}

	fmt.Print("\r                                                                                     \r")

	switch sqlQuery.Status {
	case "SUCCEEDED":
		successWriter.Write([]byte(fmt.Sprintf("SQL query completed, bytes scanned: %v, execution time: %vms", sqlQuery.BytesScanned, sqlQuery.MillisecondsExecuted)))
	case "CANCELLED", "FAILED":
		errWriter.Write([]byte("SQL query failed"))
	}

	fmt.Println()

	err = listSQLQueryResults(sqlQuery.ID)
	if err != nil {
		return err
	}

	return nil
}

func listSQLQueryResults(id string) error {
	results, err := client.GetSQLQueryResults(id)
	if err != nil {
		return err
	}

	// if len(results) > 0 {
	// 	result := results[0]
	// }

	fmt.Print(results)

	return nil
}

func listSQLQueries() error {
	request := api.NewListSQLQueriesRequest()
	request.Sort = "inserted_at.desc"
	sqlQueries, err := client.ListSQLQueries(request)
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintln(w, "id\tquery\tstatus")
	for _, sqlQuery := range sqlQueries {
		body := strings.ReplaceAll(sqlQuery.Body, "\n", " ")

		if len(body) > 75 {
			body = body[0:75] + "..."
		}

		fmt.Fprintln(w, strings.Join([]string{
			sqlQuery.ID,
			body,
			sqlQuery.Status,
		}, "\t"))
	}
	w.Flush()

	return nil
}

//
// Util
//

func waitForSQLQuery(sqlQuery *api.SQLQuery) (*api.SQLQuery, error) {
	var err error
	for {
		sqlQuery, err = client.GetSQLQuery(sqlQuery.ID)
		if err != nil {
			return nil, err
		}

		if sqlQuery.Status == "SUCCEEDED" || sqlQuery.Status == "CANCELLED" || sqlQuery.Status == "FAILED" {
			return sqlQuery, nil
		}

		s := spin.New()
		for i := 0; i < 5; i++ {
			fmt.Printf("\r%s \033[36mWaiting for query to complete, bytes scanned: %v, execution time: %vms\033[m", s.Next(), sqlQuery.BytesScanned, sqlQuery.MillisecondsExecuted)
			time.Sleep(100 * time.Millisecond)
		}
	}
}
