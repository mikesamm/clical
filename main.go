package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

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
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code:\n%v\n", authURL)

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

func clockIn() {
	// events resource: https://developers.google.com/calendar/api/v3/reference/events#resource
	fmt.Print("Clocking in\n")

	// create a temp file for clocking in
	eventStartDetails, err := os.CreateTemp("./.tmp", "newEventStartDetails-*")
	if err != nil {
		fmt.Printf("File not created: %v", err)
	}
	defer eventStartDetails.Close()

	// write to temp file: start details needed for api call
	bw, err := eventStartDetails.WriteString(time.Now().Format(time.RFC3339))
	if err != nil {
		log.Fatalf("Unable to write time to temp file: %v", err)
	}
	fmt.Printf("%v bytes written to temp file", bw)
}

func clockOut(srv *calendar.Service) {
	fmt.Print("Clocking out\n")

	// create instance of calendar.Event struct
	workEvent := &calendar.Event{
		Start:   &calendar.EventDateTime{},
		End:     &calendar.EventDateTime{},
		Summary: "",
	}

	// read from temp file, store in gcalEvent.start
	clockInFile, err := findEventTempFile("newEventStartDetails-.*")
	if err != nil {
		log.Fatalf("Failed to find temp file with start time: %v", err)
	}

	byteSliceFromTempFile, err := os.ReadFile(clockInFile)
	if err != nil {
		log.Fatalf("Failed to read temp file with start time: %v", err)
	}
	fmt.Printf("bytes from file: %v", string(byteSliceFromTempFile))
	workEvent.Start.DateTime = string(byteSliceFromTempFile)

	// store time.Now() in gcalEvent.end
	workEvent.End.DateTime = time.Now().Format(time.RFC3339)

	// assign default value to summary unless flag is specified
	// TODO:
	// 		summary comes from flags
	workEvent.Summary = "Default Summary"

	fmt.Printf("\nworkEvent start: %v", workEvent.Start.DateTime)
	fmt.Printf("\nworkEvent end: %v", workEvent.End.DateTime)
	fmt.Printf("\nworkEvent summary: %v", workEvent.Summary)

	// http request needs to have calendarId parameter
	// primary for default currently
	// calendarId := "primary"
	// newWorkEvent, err := srv.Events.Insert(calendarId, workEvent).Do()
	// if err != nil {
	// 	log.Fatalf("\nFailed to create an event on Google Calendar: %v", err)
	// }
	// fmt.Printf("new work event:\n%v", newWorkEvent)

	os.Remove(clockInFile)
}

func main() {
	ctx := context.Background()
	acceptedCommand := false
	// check for accepted commands
	for _, arg := range os.Args {
		if arg == "clockin" || arg == "ci" {
			acceptedCommand = true
		} else if arg == "clockout" || arg == "co" {
			acceptedCommand = true
		}
	}

	// check for required arguments
	if len(os.Args) <= 1 {
		log.Fatal("\n\nUsage: clical [command]\n\nCommands:\n\tclockin, ci - clock in to work\n" +
			"\tclockout, co - clock out of work\n")
	} else if !acceptedCommand {
		log.Fatal("\n\nInvalid command. Accepted commands are: clockin, ci, clockout, co")
	}

	b, err := os.ReadFile(".tmp/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// if modifying these scopes, delete your previously saved token.json
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
	if os.Args[1] == "clockin" || os.Args[1] == "ci" {
		clockIn()
	} else if os.Args[1] == "clockout" || os.Args[1] == "co" {
		clockOut(srv)
	}

}
