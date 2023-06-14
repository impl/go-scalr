package scalr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/go-cleanhttp"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/svanharmelen/jsonapi"
)

const (
	userAgent = "go-scalr"

	// DefaultAddress of Scalr.
	DefaultAddress = "https://scalr.io"
	// DefaultBasePath on which the API is served.
	DefaultBasePath = "/api/iacp/v3/"
)

var (
	// ErrWorkspaceLocked is returned when trying to lock a
	// locked workspace.
	ErrWorkspaceLocked = errors.New("workspace already locked")
	// ErrWorkspaceNotLocked is returned when trying to unlock
	// a unlocked workspace.
	ErrWorkspaceNotLocked = errors.New("workspace already unlocked")

	// ErrUnauthorized is returned when a receiving a 401.
	ErrUnauthorized = errors.New("unauthorized")

	ErrResourceNotFound = errors.New("resource not found")
)

type ResourceNotFoundError struct {
	Message string
}

func (e ResourceNotFoundError) Error() string {
	if len(e.Message) == 0 {
		return "resource not found"
	} else {
		return fmt.Sprintf(e.Message)
	}
}

func (e ResourceNotFoundError) Unwrap() error {
	return ErrResourceNotFound
}

// RetryLogHook allows a function to run before each retry.
type RetryLogHook func(attemptNum int, resp *http.Response)

// Config provides configuration details to the API client.
type Config struct {
	// The address of the Scalr API.
	Address string

	// The base path on which the API is served.
	BasePath string

	// API token used to access the Scalr API.
	Token string

	// Headers that will be added to every request.
	Headers http.Header

	// A custom HTTP client to use.
	HTTPClient *http.Client

	// RetryLogHook is invoked each time a request is retried.
	RetryLogHook RetryLogHook
}

// DefaultConfig returns a default config structure.
func DefaultConfig() *Config {
	config := &Config{
		Address:    os.Getenv("SCALR_ADDRESS"),
		BasePath:   DefaultBasePath,
		Token:      os.Getenv("SCALR_TOKEN"),
		Headers:    make(http.Header),
		HTTPClient: cleanhttp.DefaultPooledClient(),
	}

	// Set the default address if none is given.
	if config.Address == "" {
		config.Address = DefaultAddress
	}

	// Set the default user agent.
	config.Headers.Set("User-Agent", userAgent)
	// Set the default API Profile.
	config.Headers.Set("Prefer", "profile=preview")

	return config
}

// Client is the Scalr API client. It provides the basic
// connectivity and configuration for accessing the Scalr API.
type Client struct {
	baseURL           *url.URL
	token             string
	headers           http.Header
	http              *retryablehttp.Client
	retryLogHook      RetryLogHook
	retryServerErrors bool

	AccessPolicies                  AccessPolicies
	AccessTokens                    AccessTokens
	AccountUsers                    AccountUsers
	Accounts                        Accounts
	AgentPoolTokens                 AgentPoolTokens
	AgentPools                      AgentPools
	ConfigurationVersions           ConfigurationVersions
	Endpoints                       Endpoints
	EnvironmentTags                 EnvironmentTags
	Environments                    Environments
	ModuleVersions                  ModuleVersions
	Modules                         Modules
	PolicyGroupEnvironments         PolicyGroupEnvironments
	PolicyGroups                    PolicyGroups
	ProviderConfigurationLinks      ProviderConfigurationLinks
	ProviderConfigurationParameters ProviderConfigurationParameters
	ProviderConfigurations          ProviderConfigurations
	Roles                           Roles
	RunTriggers                     RunTriggers
	Runs                            Runs
	ServiceAccountTokens            ServiceAccountTokens
	ServiceAccounts                 ServiceAccounts
	SlackIntegrations               SlackIntegrations
	StateVersions                   StateVersions
	Tags                            Tags
	Teams                           Teams
	Users                           Users
	Variables                       Variables
	VcsProviders                    VcsProviders
	VcsRevisions                    VcsRevisions
	Webhooks                        Webhooks
	WebhookIntegrations             WebhookIntegrations
	WorkspaceTags                   WorkspaceTags
	Workspaces                      Workspaces
}

// NewClient creates a new Scalr API client.
func NewClient(cfg *Config) (*Client, error) {
	config := DefaultConfig()

	// Layer in the provided config for any non-blank values.
	if cfg != nil {
		if cfg.Address != "" {
			config.Address = cfg.Address
		}
		if cfg.BasePath != "" {
			config.BasePath = cfg.BasePath
		}
		if cfg.Token != "" {
			config.Token = cfg.Token
		}
		for k, v := range cfg.Headers {
			config.Headers[k] = v
		}
		if cfg.HTTPClient != nil {
			config.HTTPClient = cfg.HTTPClient
		}
		if cfg.RetryLogHook != nil {
			config.RetryLogHook = cfg.RetryLogHook
		}
	}

	// Parse the address to make sure its a valid URL.
	baseURL, err := url.Parse(config.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}

	// Only set default path if not already specified
	if baseURL.Path == "" {
		baseURL.Path = config.BasePath
	}
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	// This value must be provided by the user.
	if config.Token == "" {
		return nil, fmt.Errorf("missing API token")
	}

	// Create the client.
	client := &Client{
		baseURL:      baseURL,
		token:        config.Token,
		headers:      config.Headers,
		retryLogHook: config.RetryLogHook,
	}

	client.http = &retryablehttp.Client{
		Backoff:      retryablehttp.DefaultBackoff,
		CheckRetry:   client.retryHTTPCheck,
		ErrorHandler: retryablehttp.PassthroughErrorHandler,
		HTTPClient:   config.HTTPClient,
		RetryWaitMin: 100 * time.Millisecond,
		RetryWaitMax: 400 * time.Millisecond,
		RetryMax:     30,
	}

	// Create the services.
	client.AccessPolicies = &accessPolicies{client: client}
	client.AccessTokens = &accessTokens{client: client}
	client.AccountUsers = &accountUsers{client: client}
	client.Accounts = &accounts{client: client}
	client.AgentPoolTokens = &agentPoolTokens{client: client}
	client.AgentPools = &agentPools{client: client}
	client.ConfigurationVersions = &configurationVersions{client: client}
	client.Endpoints = &endpoints{client: client}
	client.EnvironmentTags = &environmentTag{client: client}
	client.Environments = &environments{client: client}
	client.ModuleVersions = &moduleVersions{client: client}
	client.Modules = &modules{client: client}
	client.PolicyGroupEnvironments = &policyGroupEnvironment{client: client}
	client.PolicyGroups = &policyGroups{client: client}
	client.ProviderConfigurationLinks = &providerConfigurationLinks{client: client}
	client.ProviderConfigurationParameters = &providerConfigurationParameters{client: client}
	client.ProviderConfigurations = &providerConfigurations{client: client}
	client.Roles = &roles{client: client}
	client.RunTriggers = &runTriggers{client: client}
	client.Runs = &runs{client: client}
	client.ServiceAccountTokens = &serviceAccountTokens{client: client}
	client.ServiceAccounts = &serviceAccounts{client: client}
	client.SlackIntegrations = &slackIntegrations{client: client}
	client.StateVersions = &stateVersions{client: client}
	client.Tags = &tags{client: client}
	client.Teams = &teams{client: client}
	client.Users = &users{client: client}
	client.Variables = &variables{client: client}
	client.VcsProviders = &vcsProviders{client: client}
	client.VcsRevisions = &vcsRevisions{client: client}
	client.Webhooks = &webhooks{client: client}
	client.WebhookIntegrations = &webhookIntegrations{client: client}
	client.WorkspaceTags = &workspaceTag{client: client}
	client.Workspaces = &workspaces{client: client}
	return client, nil
}

// RetryServerErrors configures the retry HTTP check to also retry
// unexpected errors or requests that failed with a server error.
func (c *Client) RetryServerErrors(retry bool) {
	c.retryServerErrors = retry
}

// retryHTTPCheck provides a callback for Client.CheckRetry which
// will retry server (>= 500) errors.
func (c *Client) retryHTTPCheck(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return c.retryServerErrors, err
	}
	if resp.StatusCode == 429 || (c.retryServerErrors && resp.StatusCode >= 500) {
		if resp.StatusCode == 429 {
			log.Printf(
				"[DEBUG] API rate limit reached for %s%s, retrying...",
				resp.Request.URL.Host, resp.Request.URL.Path,
			)
		}
		return true, nil
	}
	return false, nil
}

// newRequest creates an API request. A relative URL path can be provided in
// path, in which case it is resolved relative to the apiVersionPath of the
// Client. Relative URL paths should always be specified without a preceding
// slash.
// If v is supplied, the value will be JSONAPI encoded and included as the
// request body. If the method is GET, the value will be parsed and added as
// query parameters.
func (c *Client) newRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Authorization", "Bearer "+c.token)

	var body interface{}
	switch method {
	case "GET":
		reqHeaders.Set("Accept", "application/vnd.api+json")

		if v != nil {
			q, err := query.Values(v)
			if err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
		}
	case "DELETE", "PATCH", "POST":
		reqHeaders.Set("Accept", "application/vnd.api+json")
		reqHeaders.Set("Content-Type", "application/vnd.api+json")

		if v != nil {
			buf := bytes.NewBuffer(nil)
			if err := jsonapi.MarshalPayloadWithoutIncluded(buf, v); err != nil {
				return nil, err
			}
			body = buf
		}
	case "PUT":
		reqHeaders.Set("Accept", "application/json")
		reqHeaders.Set("Content-Type", "application/octet-stream")
		body = v
	}

	return c.createRequest(method, u.String(), body, reqHeaders)
}

func (c *Client) newJsonRequest(method, path string, v interface{}) (*retryablehttp.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Create a request specific headers map.
	reqHeaders := make(http.Header)
	reqHeaders.Set("Authorization", "Bearer "+c.token)

	var body interface{}
	reqHeaders.Set("Accept", "application/vnd.api+json")
	reqHeaders.Set("Content-Type", "application/json")

	if v != nil {
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(v); err != nil {
			return nil, err
		}
		body = buf
	}

	return c.createRequest(method, u.String(), body, reqHeaders)
}

func (c *Client) createRequest(method, url string, rawBody interface{}, reqHeaders http.Header) (*retryablehttp.Request, error) {

	req, err := retryablehttp.NewRequest(method, url, rawBody)
	if err != nil {
		return nil, err
	}

	// Set the default headers.
	for k, v := range c.headers {
		req.Header[k] = v
	}

	// Set the request specific headers.
	for k, v := range reqHeaders {
		req.Header[k] = v
	}

	return req, nil
}

// do sends an API request and returns the API response. The API response
// is JSONAPI decoded and the document's primary data is stored in the value
// pointed to by v, or returned as an error if an API error has occurred.

// If v implements the io.Writer interface, the raw response body will be
// written to v, without attempting to first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out, ctx.Err()
// will be returned.
func (c *Client) do(ctx context.Context, req *retryablehttp.Request, v interface{}) error {
	// Add the context to the request.
	req = req.WithContext(ctx)

	// Execute the request and check the response.
	resp, err := c.http.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return err
		}
	}
	defer resp.Body.Close()

	// Basic response checking.
	if err := checkResponseCode(resp); err != nil {
		return err
	}

	// Return here if decoding the response isn't needed.
	if v == nil {
		return nil
	}

	// If v implements io.Writer, write the raw response body.
	if w, ok := v.(io.Writer); ok {
		_, err = io.Copy(w, resp.Body)
		return err
	}

	// Get the value of v so we can test if it's a struct.
	dst := reflect.Indirect(reflect.ValueOf(v))

	// Return an error if v is not a struct or an io.Writer.
	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct or an io.Writer")
	}

	// Try to get the Items and Pagination struct fields.
	items := dst.FieldByName("Items")
	pagination := dst.FieldByName("Pagination")

	// Unmarshal a single value if v does not contain the
	// Items and Pagination struct fields.
	if !items.IsValid() || !pagination.IsValid() {
		return jsonapi.UnmarshalPayload(resp.Body, v)
	}

	// Return an error if v.Items is not a slice.
	if items.Type().Kind() != reflect.Slice {
		return fmt.Errorf("v.Items must be a slice")
	}

	// Create a temporary buffer and copy all the read data into it.
	body := bytes.NewBuffer(nil)
	reader := io.TeeReader(resp.Body, body)

	// Unmarshal as a list of values as v.Items is a slice.
	raw, err := jsonapi.UnmarshalManyPayload(reader, items.Type().Elem())
	if err != nil {
		return err
	}

	// Make a new slice to hold the results.
	sliceType := reflect.SliceOf(items.Type().Elem())
	result := reflect.MakeSlice(sliceType, 0, len(raw))

	// Add all of the results to the new slice.
	for _, v := range raw {
		result = reflect.Append(result, reflect.ValueOf(v))
	}

	// Pointer-swap the result.
	items.Set(result)

	// As we are getting a list of values, we need to decode
	// the pagination details out of the response body.
	p, err := parsePagination(body)
	if err != nil {
		return err
	}

	// Pointer-swap the decoded pagination details.
	pagination.Set(reflect.ValueOf(p))

	return nil
}

// ListOptions is used to specify pagination options when making API requests.
// Pagination allows breaking up large result sets into chunks, or "pages".
type ListOptions struct {
	// The page number to request. The results vary based on the PageSize.
	PageNumber int `url:"page[number],omitempty"`

	// The number of elements returned in a single page.
	PageSize int `url:"page[size],omitempty"`
}

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int `json:"current-page"`
	PreviousPage int `json:"prev-page"`
	NextPage     int `json:"next-page"`
	TotalPages   int `json:"total-pages"`
	TotalCount   int `json:"total-count"`
}

func parsePagination(body io.Reader) (*Pagination, error) {
	var raw struct {
		Meta struct {
			Pagination Pagination `json:"pagination"`
		} `json:"meta"`
	}

	// JSON decode the raw response.
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return &Pagination{}, err
	}

	return &raw.Meta.Pagination, nil
}

// checkResponseCode can be used to check the status code of an HTTP request.
func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	switch r.StatusCode {
	case 401:
		return ErrUnauthorized
	case 409:
		switch {
		case strings.HasSuffix(r.Request.URL.Path, "actions/lock"):
			return ErrWorkspaceLocked
		case strings.HasSuffix(r.Request.URL.Path, "actions/unlock"):
			return ErrWorkspaceNotLocked
		case strings.HasSuffix(r.Request.URL.Path, "actions/force-unlock"):
			return ErrWorkspaceNotLocked
		}
	}

	// Decode the error payload.
	errPayload := &jsonapi.ErrorsPayload{}
	err := json.NewDecoder(r.Body).Decode(errPayload)
	if err != nil || len(errPayload.Errors) == 0 {
		if r.StatusCode == 404 {
			return ResourceNotFoundError{}
		} else {
			return fmt.Errorf(r.Status)
		}
	}

	// Parse and format the errors.
	var errs []string
	for _, e := range errPayload.Errors {
		if e.Detail == "" {
			errs = append(errs, e.Title)
		} else {
			errs = append(errs, fmt.Sprintf("%s\n\n%s", e.Title, e.Detail))
		}
	}

	if r.StatusCode == 404 {
		return ResourceNotFoundError{
			Message: fmt.Sprint(strings.Join(errs, "\n")),
		}
	}

	if r.StatusCode == 403 {
		return fmt.Errorf(
			"The Scalr Terraform provider has been configured with an access token that lacks sufficient permissions." +
				" If you are running remotely, follow the documentation (https://docs.scalr.io/docs/scalr) on how to " +
				"enable the Scalr provider configuration in the remote workspace. " +
				"If running locally, ensure you have enough permissions to perform actions." +
				"\n Errors: " + strings.Join(errs, "\n"),
		)
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}
