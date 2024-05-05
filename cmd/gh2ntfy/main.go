package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	token := requiredEnvVar("GITHUB_TOKEN")
	err := fetchNotifications(token)
	// err := sendToNtfy()
	fmt.Printf("the: %v\n", err)
}

func requiredEnvVar(name string) string {
	val := os.Getenv(name)
	if val == "" {
		fmt.Fprintf(os.Stderr, "Required environment variable %v not set\n", name)
		os.Exit(1)
	}
	return val
}

func fetchNotifications(token string) error {
	req, err := http.NewRequest("GET", "https://api.github.com/notifications", bytes.NewReader([]byte{}))
	if err != nil {
		return err
	}

	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}

func sendToNtfy(url string) error {
	req, _ := http.NewRequest("POST", url, strings.NewReader("contents"))
	req.Header.Set("Title", "title")
	req.Header.Set("Tags", "warning,skull")
	_, err := http.DefaultClient.Do(req)
	return err
}
