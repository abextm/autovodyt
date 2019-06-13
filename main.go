package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

func main() {
	oauthConf := &oauth2.Config{
		ClientID:     os.Getenv("YOUTUBE_CLIENT_ID"),
		ClientSecret: os.Getenv("YOUTUBE_CLIENT_SECRET"),
		Scopes:       []string{youtube.YoutubeUploadScope, youtube.YoutubeReadonlyScope},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	token := getToken(oauthConf)

	client := oauthConf.Client(context.Background(), token)
	yt, err := youtube.New(client)
	if err != nil {
		panic(err)
	}

	if len(os.Args) != 1 && os.Args[1] != "-" {
		if os.Args[1] != "test" {
			fmt.Printf("unknown argument: %q\n", os.Args[1])
			os.Exit(1)
		}

		_, err := yt.Activities.List("id").Mine(true).Do()
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	streamLink := os.Getenv("STREAM_LINK")
	date := time.Now().Format("Monday January 02, 2006")
	call := yt.Videos.Insert("snippet", &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       fmt.Sprintf("Livestream %v", date),
			Description: fmt.Sprintf("Streamed live on %v at %v", date, streamLink),
			Tags:        []string{"vod"},
		},
	})
	res, err := call.Media(os.Stdin).Do()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Uploaded: %v\n", res)
}

func getToken(config *oauth2.Config) *oauth2.Token {
	fi, err := os.OpenFile("youtube_token", os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	stat, err := fi.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Size() > 0 {
		tok := &oauth2.Token{}
		err := json.NewDecoder(fi).Decode(tok)
		if err != nil {
			panic(err)
		}
		return tok
	}

	instat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (instat.Mode() & os.ModeCharDevice) == 0 {
		fmt.Printf("No OAuth token and not at terminal. Run %v test", os.Args[0])
		os.Exit(1)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to\n\t%v\nand enter the authorization code\ncode> ", authURL)
	var code string
	_, err = fmt.Scan(&code)
	if err != nil {
		panic(err)
	}

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		panic(err)
	}

	err = json.NewEncoder(fi).Encode(tok)
	if err != nil {
		panic(err)
	}

	return tok
}
