package main

import (
  "fmt"
  "io/ioutil"
)

func tail(host string, apiKey string) {
  url := fmt.Sprintf("%s%s", host, "/log_lines/search")

  resp, err := request("POST", url, nil, apiKey)
  if err != nil {
    logger.Fatal(err)
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    logger.Warn("unable to read response body")
  }
  resp.Body.Close()

  fmt.Printf("%s\n", body)
}
