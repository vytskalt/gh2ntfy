package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type Notification struct {
	Subject struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"subject"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Unread bool `json:"unread"`
}

type NotificationsResponse struct {
	Date          string
	Notifications []Notification
	PollInterval  int
}

func main() {
	githubToken := requiredEnvVar("GITHUB_TOKEN")
	ntfyURL := requiredEnvVar("NTFY_URL")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
		<-signalCh
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Listening for notifications...")

	afterDate := ""
loop:
	for {
		resp, err := fetchNotifications(ctx, githubToken, afterDate)
		if err != nil {
			log.Printf("Failed to fetch notifications: %v\n", err)
			if ctx.Err() != nil {
				break
			}

			log.Println("Will attempt again in 10s")
			select {
			case <-ctx.Done():
				break
			case <-time.After(10 * time.Second):
				continue
			}
		}

		start := time.Now()
		forwarded := 0
		for _, notification := range resp.Notifications {
			clickURL, err := fetchHTMLURL(githubToken, notification.Subject.URL)
			if err != nil {
				log.Printf("Failed to fetch HTML URL of notification source: %v\n", err)
			}

			err = sendToNTFY(ntfyURL, clickURL, &notification)
			if err != nil {
				log.Printf("Failed to forward notification to NTFY: %v\n", err)
				continue
			}

			forwarded++
		}

		if forwarded > 0 {
			log.Printf("Forwarded %v notification(s)\n", forwarded)
		}

		afterDate = resp.Date
		select {
		case <-ctx.Done():
			break loop
		case <-time.After(time.Duration(resp.PollInterval)*time.Second - time.Since(start)):
			continue
		}
	}
}

func requiredEnvVar(name string) string {
	val := os.Getenv(name)
	if val == "" {
		fmt.Fprintf(os.Stderr, "Required environment variable %v not set\n", name)
		os.Exit(1)
	}
	return val
}

func fetchHTMLURL(token, apiURL string) (string, error) {
	req, err := http.NewRequest("GET", apiURL, bytes.NewReader([]byte{}))
	if err != nil {
		return "", err
	}
	setupGitHubHeaders(req, token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil // token does not have the repo scope
	} else if resp.StatusCode != 200 {
		return "", fmt.Errorf("received unexpected status code %v when trying to get HTML URL of notification source", resp.StatusCode)
	}

	var source struct {
		HTMLURL string `json:"html_url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&source)
	if err != nil {
		return "", err
	}

	return source.HTMLURL, nil
}

func fetchNotifications(ctx context.Context, token, afterDate string) (*NotificationsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/notifications", bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}
	setupGitHubHeaders(req, token)
	req.Header.Set("If-Modified-Since", afterDate)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var notifications []Notification
	if resp.StatusCode == 200 {
		err = json.NewDecoder(resp.Body).Decode(&notifications)
		if err != nil {
			return nil, err
		}
	} else if resp.StatusCode != 304 {
		return nil, fmt.Errorf("received unexpected status code %v when trying to poll for notifications", resp.StatusCode)
	}

	date := resp.Header.Get("Date")
	pollInterval, err := strconv.Atoi(resp.Header.Get("X-Poll-Interval"))
	if err != nil {
		return nil, err
	}

	return &NotificationsResponse{
		Notifications: notifications,
		Date:          date,
		PollInterval:  pollInterval,
	}, nil
}

func sendToNTFY(ntfyURL, clickURL string, notification *Notification) error {
	content := notification.Repository.FullName + ": " + notification.Subject.Title
	req, _ := http.NewRequest("POST", ntfyURL, strings.NewReader(content))
	req.Header.Set("Title", "GitHub Notification")
	if clickURL != "" {
		req.Header.Set("Click", clickURL)
	}
	_, err := http.DefaultClient.Do(req)
	return err
}

func setupGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("User-Agent", "github.com/vytskalt/gh2ntfy")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
}
