package main

import (
  "encoding/json"
  "fmt"
)

func request(method string, path string, body interface{}) error {
  response := json.RawMessage{}

  err := client.Request(method, path, nil, response)
  if err != nil {
    return err
  }

  fmt.Printf("%s", response)

  return nil
}
