package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/savioxavier/termlink"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, save the token, then return the generated client
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := ".tmp/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Requests a token from the web, then returns the retrieved token
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then enter the "+
		"authorization code below (see README for instructions):\n\n%v\n", authURL)

	var authCode string
	// Prompt to enter the auth code
	fmt.Printf("\nEnter authCode here: ")

	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func findEventTempFile(pattern string) (string, error) {
	// read list of files in a directory
	files, err := os.ReadDir(".tmp")
	if err != nil {
		log.Fatalf("Unable to read .tmp directory: %v", err)
	}
	// compile regexp
	patternRegexp := regexp.MustCompile(pattern)
	// loop through files, match pattern
	for _, file := range files {
		if patternRegexp.MatchString(file.Name()) {
			//   return file name
			return ".tmp/" + file.Name(), nil
		}
	}
	// return empty string, error
	return "", fmt.Errorf("no temp file found matching pattern: %v", pattern)
}

func clockIn(eventSummary string) {

	// check if user is already clocked in
	eventStartTimeFile, err := findEventTempFile("newEventStartTime-.*")
	if err == nil {
		bytesFromTimeFile, err := os.ReadFile(eventStartTimeFile)
		if err != nil {
			log.Fatalf("Failed to read original clock-in file with start time: %v", err)
		}

		lastClockInDayTime, err := time.Parse(time.RFC3339, string(bytesFromTimeFile))
		if err != nil {
			log.Fatalf("Failed to parse time from RFC3339 string to ANSIC format: %v", err)
		}
		log.Fatalf("Already clocked in. Last clocked in: %v", lastClockInDayTime.Format(time.ANSIC))
	}

	// create a temp file for clocking in
	eventStartTime, err := os.CreateTemp("./.tmp", "newEventStartTime-*.txt")
	if err != nil {
		fmt.Printf("File not created: %v", err)
	}
	eventSummaryFile, err := os.CreateTemp("./.tmp", "newEventSummary-*.txt")
	if err != nil {
		fmt.Printf("File not created: %v", err)
	}
	defer eventStartTime.Close()
	defer eventSummaryFile.Close()

	// write time to temp file
	clockInTime := time.Now()
	_, err = eventStartTime.WriteString(clockInTime.Format(time.RFC3339))
	if err != nil {
		log.Fatalf("Unable to write time to temp file: %v", err)
	}

	_, err = eventSummaryFile.WriteString(eventSummary)
	if err != nil {
		log.Fatalf("Unable to write summary to temp file: %v", err)
	}

	fmt.Printf("Clocked in at: %v.\n", clockInTime.Format(time.TimeOnly))
}

func clockOut(srv *calendar.Service) {

	workEvent := &calendar.Event{
		Start:   &calendar.EventDateTime{},
		End:     &calendar.EventDateTime{},
		Summary: "",
	}

	// clock in time comes from file
	eventStartTimeFile, err := findEventTempFile("newEventStartTime-.*")
	if err != nil {
		log.Fatalf("Failed to find temp file with start time: %v", err)
	}

	bytesFromTimeFile, err := os.ReadFile(eventStartTimeFile)
	if err != nil {
		log.Fatalf("Failed to read temp file with start time: %v", err)
	}

	clockInTime := string(bytesFromTimeFile)
	workEvent.Start.DateTime = clockInTime

	// clock out time is now
	clockOutTime := time.Now()
	workEvent.End.DateTime = clockOutTime.Format(time.RFC3339)

	// summary comes from file
	eventSummaryFile, err := findEventTempFile("newEventSummary-.*")
	if err != nil {
		log.Fatalf("Failed to find temp file with summary: %v", err)
	}

	bytesFromSummaryFile, err := os.ReadFile(eventSummaryFile)
	if err != nil {
		log.Fatalf("Failed to read temp file with summary: %v", err)
	}
	workEvent.Summary = string(bytesFromSummaryFile)

	var calId string
	calIdRaw, err := os.ReadFile(".tmp/calendarId.txt")
	if err != nil {
		fmt.Println("No calendar ID provided. Event created on 'primary' calendar.")
		calId = "primary"
	} else {
		calId = string(calIdRaw)
	}

	// create full event in gcal
	newWorkEvent, err := srv.Events.Insert(calId, workEvent).Do()
	if strings.Contains(err.Error(), "Token has been expired or revoked") {
		fmt.Printf("\nWARNING: Failed to create an event on Google Calendar:"+
			"\n\t*Your Google OAuth token has expired.*"+
			"\n\tPlease clock-in again to restart the authentication token process. "+
			"\n\tYour last clock-in time was erased, but here it is for your records: %v\n", clockInTime)
		os.Remove(eventStartTimeFile)
		os.Remove(eventSummaryFile)
		os.Remove(".tmp/token.json")
		os.Exit(2)
	}
	if err != nil {
		log.Fatalf("Failed to create an event on Google Calendar: %v", err)
	}
	fmt.Printf("Clocked out at: %v\n", clockOutTime.Format(time.TimeOnly))
	fmt.Printf("See the new %s on your Google Calendar.\n", termlink.Link("work block", newWorkEvent.HtmlLink))

	os.Remove(eventStartTimeFile)
	os.Remove(eventSummaryFile)
}

func main() {
	ctx := context.Background()

	eventSummary := flag.String("s", "Work Block", "Summary (title) of event on Google Calendar.")
	flag.Parse()

	// check for required arguments
	if len(os.Args) < 2 {
		log.Fatal("\n\nUsage: clical [command]\n\nCommands:\n\tclockin, ci - clock in to work\n" +
			"\tclockout, co - clock out of work\n")
	}

	b, err := os.ReadFile(".tmp/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	// check command
	switch os.Args[len(os.Args)-1] {
	case "clockin", "ci":
		clockIn(*eventSummary)
	case "clockout", "co":
		clockOut(srv)
	default:
		log.Fatal("\n\nInvalid command. Accepted commands are: clockin, ci, clockout, co")
	}
}
