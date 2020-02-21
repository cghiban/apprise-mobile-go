package apprise

// BaseURL - base API URL
var BaseURL = "https://api.beta.theemployeeapp.com/v2"

// Client - http client for consuming the API
type Client struct {
	apiKey     string
	production bool
}

// New - creates a new apprise API client
func New(apiKey string, production bool) *Client {
	if production {
		BaseURL = "https://api.theemployeeapp.com/v2"
	}
	return &Client{apiKey: apiKey, production: production}
}
