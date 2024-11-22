package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func handler(targets []string, w http.ResponseWriter, r *http.Request) {
	// Create a channel to receive responses from targets
	ch := make(chan *http.Response, len(targets))

	// Forward the request to each target concurrently
	for _, target := range targets {
		go forwardRequest(target, r, ch)
	}

	// Track overall success/failure
	w.WriteHeader(http.StatusAccepted) // 201 Accepted
}

func forwardRequest(target string, r *http.Request, ch chan *http.Response) {
	// Create a copy of the request body
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Parse the target URL
	url, err := url.Parse(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing target URL: %v\n", err)
		ch <- nil
		return
	}

	// Create a new request with the same method and URL
	req, err := http.NewRequest(r.Method, url.String(), bytes.NewBuffer(bodyBytes)) // Use the copied body here
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating new request: %v\n", err)
		ch <- nil
		return
	}

	// Copy the headers from the original request
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Send the request to the target
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending request to target: %v\n", err)
		ch <- res
		return
	}
	fmt.Fprintf(os.Stderr, "Response from target %s: %s\n", target, res.Status)
	defer res.Body.Close()

	// Copy the response body to the channel
	ch <- res
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <target1> <target2> ...")
		return
	}

	// Get target URLs from command-line arguments
	targets := os.Args[1:]

	// Create a closure to pass targets to the handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(targets, w, r)
	})
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
