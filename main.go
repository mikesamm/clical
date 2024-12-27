package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type gcalEvent struct {
	start   calendar.EventDateTime
	end     calendar.EventDateTime
	summary string
}

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

func clockIn() {
	// events resource: https://developers.google.com/calendar/api/v3/reference/events#resource
	fmt.Print("Clocking in\n")

	// create a temp file
	eventStartDetails, err := os.CreateTemp("./.tmp", "newEventStartDetails-*")
	if err != nil {
		fmt.Printf("File not created: %v", err)
	}
	// write to temp file: start details needed for api call
	t := time.Now()
	bw, err := eventStartDetails.WriteString(t.Format(time.RFC3339))
	if err != nil {
		log.Fatalf("Unable to write time to temp file: %v", err)
	}
	fmt.Printf("%v bytes written to temp file", bw)
}

func clockOut() {
	fmt.Print("Clocking out\n")
	// http request needs to have calendarId parameter
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
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").
		Do()
	// if error, replace the token.json file, getTokenFromWeb() will be called
	if err != nil {
		log.Fatalf("unable to retrieve next ten of the user's events: %v", err)
	}

	if os.Args[1] == "clockin" || os.Args[1] == "ci" {
		clockIn()
	} else if os.Args[1] == "clockout" || os.Args[1] == "co" {
		clockOut()
	}

	fmt.Println("\n\n\nUpcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found")
	} else {
		for _, item := range events.Items {
			var date string
			startTime := item.Start.DateTime
			endTime := item.End.DateTime
			if startTime == "" || endTime == "" {
				date = item.Start.Date
				fmt.Printf("%v { Created by: %v on %v. Starts on %v }\n",
					item.Summary, item.Creator.Email, item.Created, date)
			} else {
				startTimeParsed, err := time.Parse(time.RFC3339, startTime)
				if err != nil {
					fmt.Printf("Error parsing start time: %v\n", err)
					continue
				}
				endTimeParsed, err := time.Parse(time.RFC3339, endTime)
				if err != nil {
					fmt.Printf("Error parsing end time: %v\n", err)
					continue
				}
				fmt.Printf("%v { Created by: %v on %v. Starts at %v, ends at %v }\n",
					item.Summary, item.Creator.Email, item.Created,
					startTimeParsed.Format("Mon Jan 2 15:04:05 MST 2006"),
					endTimeParsed.Format("15:04:05 MST 2006"))
			}
		}
	}

}
