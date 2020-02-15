package apprise

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// JSONTime - struct for converting time to JSON
type JSONTime struct {
	time.Time
}

func (t *JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	fmt.Println("** MarshalJSON()")
	//stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("Mon Jan _2"))
	//return []byte(stamp), nil
	return []byte(""), nil
}

func (t *JSONTime) UnmarshalJSON(buf []byte) error {
	//fmt.Println("** UnmarshalJSON()")
	tt, err := time.Parse(time.RFC3339, strings.Trim(string(buf), `"`))
	if err != nil {
		return err
	}
	t.Time = tt
	return nil
}

//Event - event struct
type Event struct {
	ID         string   `json:"_id"`
	Groups     []string `json:"accessGroups"`
	Account    string   `json:"account"`
	AllDay     bool     `json:"allday"`
	CalendarID string   `json:"calendar"`
	StartDate  JSONTime `json:"startDate"`
	EndDate    JSONTime `json:"endDate"`
	Title      string   `json:"title"`
	Notes      string   `json:"notes"`
}

type UpdateError struct {
	Message          string `json:"message"`
	Code             string `json:"code"`
	FailedValidation bool   `json:"failedValidation"`
	OriginalResponse string `json:"originalResponse"`
}

// Calendar struct
type Calendar struct {
	ID      string `json:"_id"`
	Account string `json:"account"`
	//Permissions []Permission `json:"permissions"`
	Created  JSONTime `json:"startDate"`
	Modified JSONTime `json:"endDate"`
	Title    string   `json:"title"`
}

// EventList - retrieves all the events (up to 200)
func (c *Client) EventList() ([]Event, error) {
	var events []Event
	res, err := http.Get(
		fmt.Sprintf("%s/events?limit=200&code=%s", BaseURL, c.apiKey),
	)
	if err != nil {
		log.Fatal(err)
		return events, nil
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&events); err != nil {
		return events, err
	}
	return events, nil
}

//UpdateEvent - update the given Event
func (c *Client) UpdateEvent(e Event) (Event, error) {
	var updatedEvent Event
	uri := fmt.Sprintf("%s/events/%s?code=%s", BaseURL, e.ID, c.apiKey)
	//fmt.Println(uri)
	b, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}

	var client http.Client
	req, err := http.NewRequest(
		"PUT",
		uri,
		bytes.NewBuffer(b),
	)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return updatedEvent, err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		fmt.Println(res.StatusCode, res.Status)
		var updateError UpdateError
		if err := json.NewDecoder(res.Body).Decode(&updateError); err != nil {
			log.Fatal(err)
		}
		return updatedEvent, errors.New(updateError.Message + "\n" + updateError.Code + "\n" + updateError.OriginalResponse)
	}

	if err := json.NewDecoder(res.Body).Decode(&updatedEvent); err != nil {
		return updatedEvent, err
	}
	return updatedEvent, nil
}

// DeleteEvent - delets an Event
func (c *Client) DeleteEvent(eID string) error {
	uri := fmt.Sprintf("%s/events/%s?code=%s", BaseURL, eID, c.apiKey)
	var client http.Client
	req, err := http.NewRequest("DELETE", uri, nil)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 204 {
		fmt.Println(res.StatusCode, res.Status)
		var updateError UpdateError
		if err := json.NewDecoder(res.Body).Decode(&updateError); err != nil {
			log.Fatal(err)
		}
		return errors.New(updateError.Message + "\n" + updateError.Code + "\n" + updateError.OriginalResponse)
	}

	return nil
}

// CalendarList - lists all avalable calendars
func (c *Client) CalendarList() ([]Calendar, error) {
	//fmt.Println(fmt.Sprintf("%s/events?code=%s", BaseURL, c.apiKey))
	var calendars []Calendar
	res, err := http.Get(
		fmt.Sprintf("%s/calendars?code=%s", BaseURL, c.apiKey),
	)
	if err != nil {
		log.Fatal(err)
		return calendars, nil
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&calendars)
	return calendars, err
}
