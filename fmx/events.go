package fmx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// BaseURL - base API URL
var BaseURL = "https://cshl.gofmx.com"

// JSONTime - struct for converting time to JSON
type JSONTime struct {
	time.Time
}

//UnmarshalJSON - parses FMX time and turns it into time.Time
func (t *JSONTime) UnmarshalJSON(buf []byte) error {

	loc, _ := time.LoadLocation("America/New_York")
	tt, err := time.ParseInLocation("2006-01-02T15:04:05", strings.Trim(string(buf), `"`), loc)
	if err != nil {
		return err
	}

	t.Time = tt
	return nil
}

//APIEvent - event struct for reading FMX events
type APIEvent struct {
	ID          string    `json:"id"`
	SeriesID    string    `json:"seriesID"`
	ReadURL     string    `json:"readUrl"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"` // has Location info
	AllDay      bool      `json:"allDay"`
	ClassName   string    `json:"className"`
	Start       *JSONTime `json:"start"`
	End         *JSONTime `json:"end"`
	Description string
}

//FetchEventDetails - get details for the given FMX API Event
func (e *APIEvent) FetchEventDetails() {

	res, err := http.Get(fmt.Sprintf("%s%s", BaseURL, e.ReadURL))
	if err != nil {
		log.Panicln(err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return
	}

	description := ""

	doc, err := goquery.NewDocumentFromReader(res.Body)

	// document.querySelectorAll('section.user-fieldsets div.control-group').forEach(e => {console.log(e.querySelector('.control-label').textContent, e.querySelector('.controls').textContent);})
	doc.Find("section.user-fieldsets div.control-group").Each(func(i int, s *goquery.Selection) {

		// since we get two elements (label and div), we need to use a flag
		foundField := false
		s.Find("label.control-label, div.controls").Each(func(j int, sl *goquery.Selection) {
			txt := strings.TrimSpace(sl.Text())
			lbl := sl.AttrOr("for", "")
			if strings.Contains(lbl, "CustomFields_") && strings.Contains(txt, "Description") {
				foundField = true
				return
			}
			if foundField {
				if txt != "-" { //
					if description != txt { // avoid case when short and long desc are the same
						description += txt + "\n"
					}
				}
				foundField = false
				//fmt.Println("\t", j, "\t", lbl, "\t", txt)
			}
		})
	})

	e.Description = strings.TrimSpace(description)
}

//Event - event struct
type Event struct {
	ID          string
	OccuranceID string
	Title       string
	Subtitle    string // it should be called Location
	Description string
	AllDay      bool
	Canceled    bool
	Start       *JSONTime
	End         *JSONTime
}

var reCanceled, _ = regexp.Compile("fc-event-canceled")

// Retrieve - retrieve FMX events
func Retrieve() []Event {

	eventList := []Event{}

	today := time.Now().Format("2006-01-02")
	res, err := http.Get(fmt.Sprintf("%s/calendar?date=%s&customfieldlogic=0&view=agenda&customfields=220653", BaseURL, today))
	if err != nil {
		log.Panicln(err)
		return []Event{}
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return eventList
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	doc.Find("script[data-calendar-events=\"\"]").Each(func(i int, s *goquery.Selection) {
		//fmt.Println(i, s.Text())
		jsonData := s.Text()
		err := ioutil.WriteFile("data.json", []byte(jsonData), 0644)
		if err != nil {
			log.Fatal("cant write to log file data.json:", err)
		}
		if len(jsonData) > 0 {
			var localList []APIEvent
			if err := json.Unmarshal([]byte(jsonData), &localList); err != nil {
				log.Fatalln(err)
				return
			}

			// build final list
			for _, e := range localList {

				//e.FetchEventDetails()
				/*lDesc := len(e.Description)
				if lDesc > 0 {
					if lDesc < 40 {
						fmt.Printf("** have desc: %s, [%s]\n", e.ID, e.Description)
					} else {
						fmt.Printf("** have desc: %s, [%s]", e.ID, e.Description[0:40])
					}
				}*/
				time.Sleep(800 * time.Microsecond) // 0.8th of a second

				//if !reCanceled.Match([]byte(e.ClassName)) {
				tmparray := strings.Split(e.ID, "-")
				e.ID = tmparray[2]

				tmparray = strings.Split(e.ReadURL, "/")
				fmxEvent := Event{
					ID:          e.ID,
					OccuranceID: tmparray[len(tmparray)-1],
					Title:       e.Title,
					Subtitle:    e.Subtitle,
					Description: e.Description,
					Canceled:    reCanceled.Match([]byte(e.ClassName)),
					Start:       e.Start,
					End:         e.End,
					AllDay:      e.AllDay,
				}
				eventList = append(eventList, fmxEvent)
				//}
			}

		}
	})

	return eventList
}
