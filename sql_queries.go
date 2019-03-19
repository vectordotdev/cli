package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/timberio/cli/api"
	"github.com/tj/go-spin"
)

func executeSQLQuery(query string, maxColumns int, maxColumnLength int, maxResults int) error {
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

	fmt.Println()

	err = listSQLQueryResults(sqlQuery, maxColumns, maxColumnLength, maxResults)
	if err != nil {
		return err
	}

	return nil
}

func listSQLQueryResults(sqlQuery *api.SQLQuery, maxColumns int, maxColumnLength int, maxResults int) error {
	if sqlQuery.Status == "FAILED" || sqlQuery.Status == "CANCELLED" {
		return nil
	}

	request := &api.GetSQLQueryResultsRequest{
		MaxResults: maxResults,
	}

	results, nextToken, err := client.GetSQLQueryResults(sqlQuery.ID, request)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Fprintln(errWriter, "No results")
		return nil
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	if len(results) > 0 {

		// Grab keys up to maxColumns
		var keys []string
		i := 0

		for k := range results[0] {
			if i > maxColumns {
				break
			}

			keys = append(keys, k)
			i += 1
		}

		// Print keys
		fmt.Fprint(w, "Row\t")

		for _, k := range keys {
			if len(k) > maxColumnLength {
				fmt.Fprint(w, k[0:maxColumnLength]+"...\t")
			} else {
				fmt.Fprint(w, k+"\t")
			}
		}

		fmt.Fprintln(w)

		for range keys {
			fmt.Fprint(w, "---\t")
		}

		fmt.Fprintln(w)

		for i, result := range results {
			fmt.Fprintf(w, "%v\t", i)

			var values []interface{}
			i := 0

			// Grab results up to maxColumns
			for _, v := range result {
				if i > maxColumns {
					break
				}

				values = append(values, v)
				i += 1
			}

			// Print results
			for _, v := range values {
				json, err := json.Marshal(v)
				if err != nil {
					return err
				}

				if len(json) > maxColumnLength {
					fmt.Fprintf(w, "%s\t", string(json[0:maxColumnLength])+"...")
				} else {
					fmt.Fprintf(w, "%s\t", json)
				}
			}

			fmt.Fprintln(w)
		}

		fmt.Fprintln(w)

		if nextToken != "" {
			fmt.Fprintf(warningWriter, "⚠  Only %v result shown, run `timber sql-queries download %v` to view all results\n", maxResults, sqlQuery.ID)
		} else {
			fmt.Fprintln(w, "All results shown")
		}
	}
	w.Flush()

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

		maxBodyLength := 75
		if len(body) > maxBodyLength {
			body = body[0:maxBodyLength] + "..."
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

func printSQLQueryResultsURL(sqlQuery *api.SQLQuery) error {
	if sqlQuery.ResultsURL != "" {
		fmt.Println(sqlQuery.ResultsURL)
	}

	return nil
}

func printSQLQueryInfo(sqlQuery *api.SQLQuery) error {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	statusColor := color.FgBlue

	switch sqlQuery.Status {
	case "SUCCEEDED":
		statusColor = color.FgGreen
	case "CANCELLED", "FAILED":
		statusColor = color.FgRed
	}

	fmt.Fprintf(w, "ID:\t%v\n", sqlQuery.ID)

	colorFunc := color.New(statusColor).SprintFunc()
	fmt.Fprintf(w, "Status:\t%v\n", colorFunc(sqlQuery.Status))

	if sqlQuery.FailureReason != "" {
		fmt.Fprintf(w, "Failure reason:\t%s\n", colorFunc(sqlQuery.FailureReason))
	}

	maxQueryLength := 200
	if len(sqlQuery.Body) > maxQueryLength {
		fmt.Fprintf(w, "Query:\t%v\n", sqlQuery.Body[0:maxQueryLength]+"...")
	} else {
		fmt.Fprintf(w, "Query:\t%v\n", sqlQuery.Body)
	}

	fmt.Fprintf(w, "Executed At:\t%v\n", sqlQuery.InsertedAt)

	if sqlQuery.MillisecondsExecuted > 0 {
		fmt.Fprintf(w, "Duration:\t%vms\n", sqlQuery.MillisecondsExecuted)
	}

	if sqlQuery.BytesScanned > 0 {
		fmt.Fprintf(w, "Scanned:\t%v bytes\n", sqlQuery.BytesScanned)
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
