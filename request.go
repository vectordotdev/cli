package main

import (
	"encoding/json"
	"fmt"
)

func request(method string, path string, body interface{}) error {
	response := json.RawMessage{}

	err := client.Request(method, path, nil, nil, &response)
	if err != nil {
		return err
	}

	json, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		logger.Fatal(err)
	}

	fmt.Printf("%s\n", json)

	return nil
}
