package main

import (
  "fmt"
  "net/http"
  "time"

  "github.com/hashicorp/go-retryablehttp"
)

var httpClient = retryablehttp.NewClient()
var userAgent = fmt.Sprintf("timber-cli/%s", version)

func init() {
  httpClient.HTTPClient.Timeout = 10 * time.Second
  httpClient.Logger = logger
}

func request(method string, endpoint string, rawBody interface{}, apiKey string) (*http.Response, error) {
  req, err := retryablehttp.NewRequest(method, endpoint, rawBody)
  if err != nil {
    logger.Fatal(err)
  }

  req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
  req.Header.Add("User-Agent", userAgent)

  return httpClient.Do(req)
}
