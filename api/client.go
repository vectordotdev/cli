package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var userAgent = fmt.Sprintf("timber-cli/%s", "0.1.0")

type Client struct {
	APIKey string
	Host   string

	httpClient *retryablehttp.Client
}

type Logger interface {
	Printf(string, ...interface{})
}

func NewClient(host string, apiKey string) *Client {
	httpClient := retryablehttp.NewClient()

	httpClient.HTTPClient.Timeout = 10 * time.Second

	return &Client{
		APIKey: apiKey,
		Host:   host,

		httpClient: httpClient,
	}
}

func (c *Client) SetLogger(l Logger) {
	c.httpClient.Logger = l
}

type searchRequest struct {
	ApplicationIds      []string  `json:"application_ids"`
	DatetimeGreaterThan time.Time `json:"dt_gt"`
	Limit               int       `json:"limit"`
	Query               string    `json:"query"`
	Sort                string    `json:"sort"` // TODO maybe make this an "enum"
}

func (c *Client) Search(appIds []string, datetimeGreaterThan time.Time, query string) ([]*LogLine, error) {
	limit := 250

	response := struct {
		LogLines []*LogLine `json:"data"`
	}{
		LogLines: make([]*LogLine, 0, limit),
	}

	err := c.request("POST", "/log_lines/search", searchRequest{
		ApplicationIds:      appIds,
		DatetimeGreaterThan: datetimeGreaterThan,
		Query:               query,
		Limit:               limit,
		Sort:                "dt.desc",
	}, &response)
	if err != nil {
		return nil, err
	}

	return response.LogLines, nil
}

func (c *Client) ListApplications() ([]*Application, error) {
	response := struct {
		Applications []*Application `json:"data"`
	}{}

	err := c.request("GET", "/applications", nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Applications, nil
}

func (c *Client) request(method string, path string, requestStruct interface{}, responseStruct interface{}) error {
	url := fmt.Sprintf("%s%s", c.Host, path)

	var rawBody interface{}
	rawBody = nil
	if requestStruct != nil {
		b, err := json.Marshal(requestStruct)
		if err != nil {
			return err
		}

		rawBody = bytes.NewBuffer(b)
	}

	req, err := retryablehttp.NewRequest(method, url, rawBody)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))
	req.Header.Add("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if responseStruct != nil {
		err = json.NewDecoder(resp.Body).Decode(responseStruct)
		if err != nil {
			return err
		}
	}

	return nil
}