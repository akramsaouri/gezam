package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/browser"
)

type Track struct {
	name   string
	artist string
}

type GeniusResponse struct {
	Data struct {
		Hits []struct {
			Index     string `json:"string"`
			HitResult struct {
				FullTitle         string `json:"full_title"`
				Title             string `json:"title"`
				TitleWithFeatured string `json:"title_with_featured"`
				Path              string `json:"path"`
				PrimaryArtist     struct {
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

func cleanupString(name string) string {
	re := regexp.MustCompile(`\([^)]*\)`)
	return string(re.ReplaceAll([]byte(name), []byte("")))
}

func openTrack(track Track) {
	// fmt.Println("q: ", cleanupString(track.name)+" "+track.artist)
	url := fmt.Sprintf("https://genius.com/api/search?q=%s", url.QueryEscape(cleanupString(track.name)+" "+track.artist))
	// fmt.Println("url: ", url)
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		sentry.CaptureException(err)
	}
	dec := json.NewDecoder(resp.Body)
	var response GeniusResponse
	err = dec.Decode(&response)
	if err != nil {
		sentry.CaptureException(err)
	}
	var hitResultPath string
	for _, hit := range response.Data.Hits {
		if strings.Contains(strings.ToLower(track.name), strings.ToLower(hit.HitResult.Title)) &&
			strings.Contains(strings.ToLower(track.artist), strings.ToLower(hit.HitResult.PrimaryArtist.Name)) {
			hitResultPath = hit.HitResult.Path
			break
		}
	}
	if hitResultPath != "" {
		geniusURL := fmt.Sprintf("https://genius.com%s", hitResultPath)
		browser.OpenURL(geniusURL)
	}
}

func getCurrentTrack() Track {
	cmd := exec.Command("/usr/bin/osascript", "-e", `
	if application "Spotify" is running then
		tell application "Spotify"
			return (get name of current track) & "||" & (get artist of current track)
		end tell
	else
		return "ERR_SPOTIFY_NOT_RUNNING"
	end if
	return`)
	cmd.Stdin = strings.NewReader("")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		sentry.CaptureException(err)
	}
	result := strings.TrimSpace(out.String())
	if strings.Contains(result, "ERR_SPOTIFY_NOT_RUNNING") {
		return Track{
			name: "gezam",
		}
	}
	chunks := strings.Split(strings.TrimSpace(out.String()), "||")
	if len(chunks) > 0 {
		var artist string
		if len(chunks) > 1 {
			artist = chunks[1]
		}
		return Track{
			name:   chunks[0],
			artist: artist,
		}
	}
	return Track{
		name: "gezam",
	}
}

func formatTrack(track Track) string {
	return fmt.Sprintf("%s - %s", track.name, track.artist)
}

func renderCurrentTrack() {
	for {
		menuet.App().SetMenuState(&menuet.MenuState{
			Title: formatTrack(getCurrentTrack()),
		})
		time.Sleep(time.Second)
	}
}

func menu() []menuet.MenuItem {
	return []menuet.MenuItem{
		{
			Text:     "Read lyrics",
			FontSize: 14,
			Clicked: func() {
				currentTrack := getCurrentTrack()
				if currentTrack.name != "gezam" {
					openTrack(getCurrentTrack())
				}
			},
		},
	}
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://0f00351e90934e3c981745dcef5902c2@o292706.ingest.sentry.io/5674391",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)
	go renderCurrentTrack()
	menuet.App().Label = "com.github.akramsaouri.gezam"
	menuet.App().Children = menu
	menuet.App().RunApplication()
}
