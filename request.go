package main

import (
  "fmt"
)

func request(method string, path string, body interface{}) error {
  response := make(map[string]interface{})

  err := client.Request(method, path, nil, response)
  if err != nil {
    return err
  }

  fmt.Printf("%v", response)

  return nil
}
