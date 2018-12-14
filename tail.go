package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "time"
)

func tail(host string, apiKey string) {
  url := fmt.Sprintf("%s%s", host, "/log_lines/search")

  resp, err := request("POST", url, nil, apiKey)
  if err != nil {
    logger.Fatal(err)
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    logger.Fatal(err)
  }
  resp.Body.Close()

  m := make(map[string]interface{})

  err = json.Unmarshal(body, &m)
  if err != nil {
    logger.Fatal(err)
  }

  fmt.Printf("%v\n", m["data"])

  time.Sleep(2 * time.Second)

  tail(host, apiKey)
}
