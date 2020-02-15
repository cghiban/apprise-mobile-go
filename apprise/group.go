package apprise

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Group - struct for the gruoup
type Group struct {
	ID      string   `json:"_id"`
	Account string   `json:"account"`
	Created JSONTime `json:"created"`
	Name    string   `json:"name"`
}

// GroupList - fetches all the groups
func (c *Client) GroupList() ([]Group, error) {
	var groups []Group
	url := fmt.Sprintf("%s/groups?limit=200&code=%s", BaseURL, c.apiKey)
	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return groups, nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return groups, errors.New("Error: Invalid Grant Code")
	}

	err = json.NewDecoder(res.Body).Decode(&groups)
	return groups, err
}
