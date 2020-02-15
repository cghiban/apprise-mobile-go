package apprise

// BaseURL - base API URL
const BaseURL = "https://api.theemployeeapp.com/v2"

// Client - http client for consuming the API
type Client struct {
	apiKey     string
	production bool
}

// New - creates a new apprise API client
func New(apiKey string, production bool) *Client {
	return &Client{apiKey: apiKey, production: production}
}
