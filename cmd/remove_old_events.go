package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"../apprise"
)

const twoWeeks = 2 * 7 * 24 * time.Hour

var apiKey string
var production bool

func init() {
	apiKey = os.Getenv("APIKEY")
	if apiKey == "" {
		log.Fatalln("Apprise APIKEY is not set")
	}

	prod := os.Getenv("PRODUCTION")
	if len(prod) > 0 {
		production = true
	}
}

func main() {
	var api = apprise.New(apiKey, production)

	events, err := api.EventList()
	if err != nil {
		log.Fatal(err)
	}

	twoWeeksAgo := time.Now().UTC().Add(-twoWeeks)
	fmt.Println("Removing events older than: ", twoWeeksAgo)
	for _, e := range events {
		if e.StartDate.Time.Before(twoWeeksAgo) {
			fmt.Println("Will delete:", e.ID, e.StartDate)
			if err := api.DeleteEvent(e.ID); err != nil {
				fmt.Println(err)
			}
		}
	}
}
