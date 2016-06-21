package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mrjones/oauth"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

const (
	authURI         = "https://accounts.google.com/o/oauth2/auth"
	tokenURI        = "https://accounts.google.com/o/oauth2/token"
	redirectURI     = "urn:ietf:wg:oauth:2.0:oob"
	twitterEndpoint = "https://api.twitter.com/1.1/statuses/update.json"
)

var (
	sess                     = session.New()
	svc                      = s3.New(sess)
	bucketName               = aws.String(os.Getenv("AWS_S3_BUCKET_NAME"))
	twitterConsumerKey       = os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret    = os.Getenv("TWITTER_CONSUMER_SECRET")
	twitterAccessToken       = os.Getenv("TWITTER_ACCESS_TOKEN")
	twitterAccessTokenSecret = os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	consumer                 = oauth.NewConsumer(
		twitterConsumerKey,
		twitterConsumerSecret,
		oauth.ServiceProvider{},
	)
	oauthAccessToken = &oauth.AccessToken{
		Token:  twitterAccessToken,
		Secret: twitterAccessTokenSecret,
	}
	client, _ = consumer.MakeHttpClient(oauthAccessToken)
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile := "calendar-go-quickstart.json"
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	params := &s3.GetObjectInput{
		Bucket: bucketName,
		Key:    aws.String(file),
	}
	res, err := svc.GetObject(params)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(res.Body).Decode(t)
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)

	b, err := json.Marshal(token)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	params := &s3.PutObjectInput{
		Bucket: bucketName,
		Key:    aws.String(file),
		Body:   bytes.NewReader(b),
	}
	_, err = svc.PutObject(params)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
}

func formatEvents(events []*calendar.Event) string {
	tmpl := `参加予定のイベントです！
{{range .}}{{.Summary}} ({{if .Start.DateTime}}{{.Start.DateTime}}{{else}}{{.Start.Date}}{{end}})
{{end}}`
	t := template.Must(template.New("t").Parse(tmpl))

	b := &bytes.Buffer{}
	err := t.Execute(b, events)
	if err != nil {
		log.Fatalf("Unable to execute template %v", err)
	}

	return b.String()
}

func updateTwitter(params map[string]string) {
	vals := url.Values{}
	for k, v := range params {
		vals.Add(k, v)
	}

	req, err := http.NewRequest("POST", twitterEndpoint, strings.NewReader(vals.Encode()))
	if err != nil {
		log.Fatalf("Unable to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to request twitter: %v", err)
	}

	log.Printf("Response: %v", res.Status)
}

func main() {
	ctx := context.Background()

	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  redirectURI,
		Scopes:       []string{calendar.CalendarReadonlyScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURI,
			TokenURL: tokenURI,
		},
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	from := time.Now().Format(time.RFC3339)
	to := time.Now().AddDate(0, 0, 15).Format(time.RFC3339)
	events, err := srv.Events.List("primary").ShowDeleted(false).SingleEvents(true).
		TimeMin(from).TimeMax(to).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	if len(events.Items) > 0 {
		updateTwitter(map[string]string{
			"status": formatEvents(events.Items),
		})
	} else {
		log.Printf("No upcoming events found.\n")
	}
}
