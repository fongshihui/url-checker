package main

import (
	"net/http"
	"time"
)

type result struct {
	url      string
	status   int
	err      error
	attempts int
	duration time.Duration
}

func worker(id int, client *http.Client, retries int, jobs <-chan string, results chan<- result) {
	for raw := range jobs {
		start := time.Now()
		status, attempts, err := checkURL(client, raw, retries)
		results <- result{
			url:      raw,
			status:   status,
			err:      err,
			attempts: attempts,
			duration: time.Since(start),
		}
	}
}
