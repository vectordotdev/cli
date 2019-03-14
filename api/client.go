package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
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

	httpClient.Logger = nil

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

//
// Logs
//

type searchRequest struct {
	ApplicationIds []string  `json:"application_ids"`
	DtGt           time.Time `json:"dt_gt"`
	Limit          int       `json:"limit"`
	Query          string    `json:"query"`
	Sort           string    `json:"sort"` // TODO maybe make this an "enum"
}

func NewSearchRequest() *searchRequest {
	request := &searchRequest{}
	request.Limit = 250
	request.Sort = "dt.desc"
	return request
}

func (c *Client) Search(request *searchRequest) ([]*LogLine, error) {
	response := struct {
		RawLines []*json.RawMessage `json:"data"`
	}{
		make([]*json.RawMessage, 0, request.Limit),
	}

	err := c.Request("POST", "/log_lines/search", nil, request, &response)
	if err != nil {
		return nil, err
	}

	logLines := make([]*LogLine, len(response.RawLines))
	for i, rawLine := range response.RawLines {
		// unmarshal twice, once to fill structured fields, once to unmarshal the unknown fields
		// TODO it'd be better to only unmarshal once
		logLine := &LogLine{}

		if err := json.Unmarshal(*rawLine, logLine); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(*rawLine, &logLine.Fields); err != nil {
			return nil, err
		}

		logLines[i] = logLine
	}

	return logLines, nil
}

//
// Sources
//

func (c *Client) ListSources() ([]*Application, error) {
	response := struct {
		Applications []*Application `json:"data"`
	}{}

	err := c.Request("GET", "/applications", nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Applications, nil
}

//
// Organizations
//

func (c *Client) GetOrganization(id string) (*Organization, error) {
	response := struct {
		Organization *Organization `json:"data"`
	}{}

	err := c.Request("GET", path.Join("/organizations/", id), nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Organization, nil
}

func (c *Client) ListOrganizations() ([]*Organization, error) {
	response := struct {
		Organizations []*Organization `json:"data"`
	}{}

	err := c.Request("GET", "/organizations", nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Organizations, nil
}

//
// Saved Views
//

func (c *Client) GetSavedView(id string) (*SavedView, error) {
	response := struct {
		SavedView *SavedView `json:"data"`
	}{}

	err := c.Request("GET", path.Join("/saved_views", id), nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.SavedView, nil
}

func (c *Client) ListSavedViews() ([]*SavedView, error) {
	response := struct {
		SavedViews []*SavedView `json:"data"`
	}{}

	err := c.Request("GET", "/saved_views?type=CONSOLE", nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.SavedViews, nil
}

//
// SQL Queries
//

func (c *Client) CreateSQLQuery(organizationID string, body string) (*SQLQuery, error) {
	request := struct {
		OrganizationID string `json:"organization_id"`
		Body           string `json:"body"`
	}{
		OrganizationID: organizationID,
		Body:           body,
	}

	response := struct {
		SQLQuery *SQLQuery `json:"data"`
	}{}

	err := c.Request("POST", "/sql_queries", nil, &request, &response)
	if err != nil {
		return nil, err
	}

	return response.SQLQuery, nil
}

func (c *Client) GetSQLQuery(id string) (*SQLQuery, error) {
	response := struct {
		SQLQuery *SQLQuery `json:"data"`
	}{}

	err := c.Request("GET", path.Join("/sql_queries", id), nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.SQLQuery, nil
}

func (c *Client) GetSQLQueryResults(id string) ([]interface{}, error) {
	response := struct {
		Results []interface{} `json:"data"`
	}{}

	err := c.Request("GET", path.Join("/sql_queries", id, "results"), nil, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

type ListSQLQueriesRequest struct {
	Limit int    `json:"limit"`
	Sort  string `json:"sort"` // TODO maybe make this an "enum"
}

func NewListSQLQueriesRequest() *ListSQLQueriesRequest {
	request := &ListSQLQueriesRequest{}
	request.Limit = 25
	request.Sort = "inserted_at.desc"
	return request
}

func (c *Client) ListSQLQueries(request *ListSQLQueriesRequest) ([]*SQLQuery, error) {
	response := struct {
		SQLQueries []*SQLQuery `json:"data"`
	}{}

	query := url.Values{}

	if request.Limit > 0 {
		query.Set("limit", strconv.Itoa(request.Limit))
	}

	if request.Sort != "" {
		query.Set("sort", request.Sort)
	}

	err := c.Request("GET", "/sql_queries", &query, nil, &response)
	if err != nil {
		return nil, err
	}

	return response.SQLQueries, nil
}

//
// Util
//

func (c *Client) Request(method string, path string, query *url.Values, requestStruct interface{}, responseStruct interface{}) error {
	if c.Host == "" {
		return errors.New("A host is required to make a request to the Timber API")
	}

	url := fmt.Sprintf("%s%s", c.Host, path)

	if query != nil {
		url = url + "?" + query.Encode()
	}

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

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if responseStruct != nil {
			err = json.NewDecoder(resp.Body).Decode(responseStruct)
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		response := struct {
			Error  *Error   `json:"error"`
			Errors []*Error `json:"errors"`
		}{}

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return err
		}

		error := response.Error

		if error == nil && len(response.Errors) > 0 {
			error = response.Errors[0]
		}

		return &ServiceError{StatusCode: resp.StatusCode, ErrorStruct: error}
	}
}
