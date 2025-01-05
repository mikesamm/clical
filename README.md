<div align="center">
  <h1>>🗓️ clical</h1>
</div>

<br />

## About the Project

**clical** is a command line application, written in Go, that facilitates tracking your hours worked in Google Calendar. When you clock in then clock out, an event is created, yielding a visual representation of your hours worked. 😎

<!-- GETTING STARTED -->
## Getting Started

### Requirements
1. A Google account.
2. A Google Cloud project. I recommend [setting up your project](#authenticating-for-the-google-apis) first.
3. Go installed on your machine, docs [here](https://go.dev/doc/install).

<!-- Installation -->
### Installation

<!-- Run Locally -->
#### Build from source, run locally

0. Clone the project

```bash
  git clone https://github.com/mikesamm/clical.git
```

1. Go to the project's root directory

```bash
  cd clical
```
2. [Set calendar ID](#find-and-set-your-google-calendar-id) to your desired calendar (or don't. **clical** will default to your 'primary' calendar).

3. Add a `credentials.json` file to `/.tmp`. These credentials should be the `.json` file you downloaded from your [Google Cloud project](#part-2-creating-an-auth-client-and-token).

4. Build **clical**
```bash
go build clical
```

## Usage

```
./clical [OPTIONS] [COMMAND]
```

1. Run `./clical ci` or `./clical clockin`.
    1. On the first run, it will prompt you to follow an authentication link. Click through the OAuth consent process. This process is directly connected to your [Google Cloud project](#authenticating-for-the-google-apis).
    2. Once you've continued as far as you can and it appears that the redirect failed, you're in the right place.
    3. In the url, find the `code=` query parameter. Copy the entire value for that parameter up to and NOT including the next `&` (you'll likely see `&scope=` as the next parameter).
    4. You should now have a very long string of characters. Paste that in your terminal at the `Enter authCode here:` prompt and press enter.
    5. Your token should be saved to a file in the `/.tmp` directory, and you should see a successful clock in confirmation with the current time.
2. When you're ready, run `./clical co` or `./clical clockin` and you will see a confirmation with a link to the event you just created.

**currently clical can only run from its root directory*

### Examples

```
./clical ci                  // clock in with default event summary
./clical -s "coffe shop" ci  // clock in with custom summary
./clical co                  // clock out
```

### Options
```
-s        // create a custom event summary
```
**A summary in Google Calendar is the title (what you see on the actual calendar block).*
## Planned features
- `status` command to check whether if you're currently clocked in or not.
- `total` command to see your total hours for a given time period (with options for day, week month, year, custom range).

## Known Bugs
- [issue #1](https://github.com/mikesamm/clical/issues/1#issue-2769463062) - clocking in twice breaks clock out.

<!-- License -->
## License

Distributed under the MIT License. See LICENSE.txt for more information.


<!-- Contact -->
## Contact

Mike Sammartino - mike@mikesammartino.com

Project Link: [https://github.com/mikesamm/clical](https://github.com/mikesamm/clical)

 ## Appendix

 ### Find and Set your Google Calendar ID

1. Open Google Calendar.
2. Find the list of calendars on the bottom left.
3. Hover over your desired calendar, click the three dots that appear, then click "Settings and sharing".
4. Under "Integrate calendar" you'll see the calendar id. Copy the complete string and replace the placeholder text in `.tmp/calendar.txt.example`.
5. Truncate `calendar.txt.example` to `calendar.txt`.
6. **clical** will now create events in this calendar.

### Authenticating for the Google APIs

**clical** needs to be granted permission to integrate with Google APIs for your account before it can work properly.

These initial setup steps can look a little intimidating, but they're completely safe and similar to the setup currently needed for most non-commercial projects that integrate with Google APIs (for example, [gmailctl] or [Home Assistant](https://www.home-assistant.io/integrations/google_assistant/)).

[gmailctl]: https://github.com/mbrt/gmailctl

#### Part 1: Setting up your Google "project"

To generate an OAuth token, you first need a placeholder "project" in Google Cloud Console. Create a new project if you don't already have a generic one you can reuse.

1. [Create a New Project](https://console.developers.google.com/projectcreate) within the Google
   developer console
  1. (You can skip to the next step if you already have a placeholder project you can use)
  2. Pick any project name. Example: "Placeholder Project 12345"
  3. Click the "Create" button.
2. [Enable the Google Calendar API](https://console.developers.google.com/apis/api/calendar-json.googleapis.com/)
  1. Click the "Enable" button.

#### Part 2: Creating an auth client and token

Once you have the project with Calendar API enabled, you need a way for **clical** to request permission to use it on a user's account.

1. [Create OAuth2 consent screen](https://console.developers.google.com/apis/credentials/consent/edit;newAppInternalUser=false) for a "UI/Desktop Application".
   1. Fill out required App information section
      1. Specify App name. Example: "clical"
      2. Specify User support email. Example: your@gmail.com
   2. Fill out required Developer contact information
      1. Specify Email addresses. Example: your@gmail.com
   3. Click the "Save and continue" button.
   4. Scopes:
      1. add the `.../auth/calendar.events` scope.
      2. click the "Save and continue" button.
   5. Test users
      1. Add your@gmail.com
      2. Click the "Save and continue" button.
2. [Create OAuth Client ID](https://console.developers.google.com/apis/credentials/oauthclient)
   1. Specify Application type: Desktop app.
   2. Click the "Create" button.
3. Grab your newly created credentials. You'll see Client ID (in the form "xxxxxxxxxxxxxxx.apps.googleusercontent.com") and Client Secret either on the Credentials page or Clients page.
4. Click the download button next to the Client Secret. You've downloaded the `.json` file containing your credentials.
5. Download as `credentials.json` in the `clical/.tmp` directory or create a `credentials.json` file in `/.tmp` and copy the contents of the downloaded file.
