package main

import (
	//"encoding/json"

	"apprise/apprise"
	"apprise/fmx"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FMXEvent - structure for locally stored events
type FMXEvent struct {
	FmxID       string    `db:"fmx_id"`
	OccuranceID string    `db:"occurance_id"`
	APIID       string    `db:""api_id`
	Title       string    `db:"title"`
	Notes       string    `db:"notes"`
	StartDate   time.Time `db:"startdate"`
	EndDate     time.Time `db:"enddate"`
	AllDay      bool      `db:"all_day"`
	Canceled    bool      `db:"canceled"`
}

func createTable(db *sql.DB) {
	sql_table := `
		CREATE TABLE IF NOT EXISTS fmxevents (
			fmx_id TEXT NOT NULL,
			occurance_id TEXT NOT NULL,
			api_id TEXT NOT NULL,
			title TEXT,
			notes TEXT,
			startdate DATETIME,
			enddate DATETIME,
			all_day BOOLEAN,
			canceled BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS eventIndex ON fmxevents (fmx_id, occurance_id);
	CREATE INDEX IF NOT EXISTS apiIndex ON fmxevents (api_id);
	`
	res, err := db.Exec(sql_table)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func initDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic("db nil")
	}
	createTable(db)
	return db
}

func findEvent(db *sql.DB, fmxID, occuranceID string) *FMXEvent {
	sql_findevent := `
	SELECT fmx_id, occurance_id, api_id, title, notes, startdate, enddate, canceled, all_day
	FROM fmxevents
	WHERE fmx_id = ? AND occurance_id = ?`

	row := db.QueryRow(sql_findevent, fmxID, occuranceID)
	var e FMXEvent
	err := row.Scan(&e.FmxID, &e.OccuranceID, &e.APIID, &e.Title, &e.Notes, &e.StartDate, &e.EndDate, &e.Canceled, &e.AllDay)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal(err)
	}

	return &e
}

func storeEvent(db *sql.DB, e *FMXEvent) {
	sql_addevent := `
	INSERT OR REPLACE INTO fmxevents (
		fmx_id, occurance_id, api_id, title, notes, startdate, enddate, canceled, all_day, created_at
	) values(?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	stmt, err := db.Prepare(sql_addevent)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	//fmt.Printf("%q\n", e)

	res, err2 := stmt.Exec(e.FmxID, e.OccuranceID, e.APIID, e.Title, e.Notes, e.StartDate, e.EndDate, e.Canceled, e.AllDay)
	if err2 != nil {
		panic(err2)
	}
	rowNum, _ := res.RowsAffected()
	fmt.Println(" -- added to DB: ", rowNum, e.APIID)
}

func updateEvent(db *sql.DB, e *FMXEvent) {
	sql_updateevent := `
	UPDATE fmxevents 
	SET api_id = ?, title = ?, notes = ?, startdate = ?, enddate = ?, canceled = ?, all_day = ?, updated_at = CURRENT_TIMESTAMP
	WHERE fmx_id = ? AND occurance_id = ?
	`
	stmt, err := db.Prepare(sql_updateevent)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(e.APIID, e.Title, e.Notes, e.StartDate, e.EndDate, e.Canceled, e.AllDay, e.FmxID, e.OccuranceID)
	if err != nil {
		panic(err)
	}
	rowNum, _ := res.RowsAffected()
	fmt.Println(" -- updated: ", rowNum, e.APIID)
}

var apiKey, dbPath string
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

	dbPath = os.Getenv("SQLITE_DB")
	if dbPath == "" {
		log.Fatalln("SQLITE_DB is not set")
	}
}

var api *apprise.Client

func publishEvent(db *sql.DB, e fmx.Event) error {
	startDate := apprise.JSONTime{Time: e.Start.Time.UTC()}
	var endDate apprise.JSONTime
	if e.End == nil {
		loc, _ := time.LoadLocation("America/New_York")
		endDateString := e.Start.Time.Format("2006-01-02") + "T23:59:59"
		endDateZZ, _ := time.ParseInLocation("2006-01-02T15:04:05", endDateString, loc)
		endDate = apprise.JSONTime{Time: endDateZZ.UTC()}
		fmt.Printf("%v -- %v -- %v", startDate, endDateZZ, endDate)
	} else {
		endDate = apprise.JSONTime{Time: e.End.Time.UTC()}
	}

	// create new event
	eventData := apprise.Event{
		Groups:     []string{"5d60171f68c5ce00e1506dcb", "5d6429decbe4cc00e19375fb"},
		AllDay:     e.AllDay,
		CalendarID: os.Getenv("CALENDAR"),
		StartDate:  startDate,
		EndDate:    endDate,
		Title:      e.Title,
		Notes:      strings.TrimSpace(e.Subtitle + "\n" + e.Description),
	}

	aEvent, err := api.CreateEvent(eventData)
	if err != nil {

		j, err := json.Marshal(eventData)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%+v\n", eventData)
		fmt.Printf("%+v\n", string(j))

		log.Println(err)
		return err
	}
	//fmt.Printf("\nadded event: %+v\n", aEvent)
	//fmt.Printf("* added event: %s\t%s\t%s\n", aEvent.ID, aEvent.StartDate, aEvent.Title)
	//add aEvent into our DB

	fmxEvent := &FMXEvent{
		FmxID:       e.ID,
		OccuranceID: e.OccuranceID,
		APIID:       aEvent.ID,
		Title:       aEvent.Title,
		Notes:       aEvent.Notes,
		StartDate:   aEvent.StartDate.Time,
		EndDate:     aEvent.EndDate.Time,
		AllDay:      aEvent.AllDay,
		Canceled:    false,
	}

	storeEvent(db, fmxEvent)

	return nil
}

func isModified(dbevent *FMXEvent, fe *fmx.Event) bool {

	/*
		fmt.Printf("%+v\n", dbevent)
		fmt.Printf("~~~~~~~~~~~~~~~~~~~~~\n")
		fmt.Printf("%+v\n", fe)
	*/

	if dbevent.StartDate.Local().String() != fe.Start.Time.String() {
		fmt.Printf(" >> Start: %v -> %v\n", dbevent.StartDate.Local(), fe.Start.Time)
		return true
	}

	if fe.End != nil && dbevent.EndDate.Local().String() != fe.End.Time.String() {
		fmt.Printf(" >> End: %v -> %v\n", dbevent.EndDate.Local(), fe.End.Time)
		return true
	}

	if dbevent.Title != fe.Title {
		fmt.Printf(" >> Title: %s -> %s\n", dbevent.Title, fe.Title)
		return true
	}
	description := strings.TrimSpace(fe.Subtitle + "\n" + fe.Description)
	if dbevent.Notes != description {
		fmt.Printf(" >> Notes: {%s} -> {%s}\n", dbevent.Notes, description)
		return true
	}

	return false
}

// details from here:
// https://cshl.gofmx.com/scheduling/requests/2201895/occurrences/5209400
//
func main() {

	api = apprise.New(apiKey, production)
	db := initDB(dbPath)

	fmxEvents := fmx.Retrieve()
	for _, fe := range fmxEvents {
		dbevent := findEvent(db, fe.ID, fe.OccuranceID)
		if fe.Canceled {
			// check is already added
			// if added, check if canceled
			//	 if canceled ,remove from AppRise AND remove from DB

			if dbevent != nil && dbevent.APIID != "" {
				fmt.Println("XX", fe.ID, "\t", fe.Start, "\t", fe.AllDay, "\t", fe.Title, fe.Subtitle)

				err := api.DeleteEvent(dbevent.APIID)
				if err != nil && !strings.Contains(err.Error(), "No object found with") {
					fmt.Println("Error deleting api event: ", err)
				} else {
					fmt.Println("** Deleted: ", dbevent.APIID, dbevent.Title)
					// update event, remove its api_id
					dbevent.APIID = ""
					updateEvent(db, dbevent)
				}
			}
			// do not remove this line
			continue
		}

		if dbevent != nil {
			// we also need to check if event was modified
			fmt.Printf("found %s,%s,%s\t%s\t%s\n", dbevent.FmxID, dbevent.OccuranceID,
				dbevent.APIID, dbevent.StartDate.Local().Format("1/2 3:4"), dbevent.Title,
			)
			if isModified(dbevent, &fe) {
				fmt.Println("---- ^^^ modified ------------------------------------")
			}
		} else {
			fmt.Println("ADDING", fe.ID, "\t", fe.OccuranceID, "\t", fe.Start, "\t", fe.Title, fe.Subtitle)
			if err := publishEvent(db, fe); err != nil {
				fmt.Printf("---- error publishing %q:\n%q\n", fe, err)
			}
			//break
		}

	}
}
