package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
)

func run(cfg config, urls []string) (int, int, error) {
	// Validate runtime configuration before doing any work.
	if err := validateConfig(cfg); err != nil {
		return 0, 0, err
	}

	// Shared HTTP client for all workers.
	client := &http.Client{Timeout: cfg.timeout}
	// Unbuffered channels provide backpressure between producer, workers, and consumer.
	jobs := make(chan string)
	results := make(chan result)

	// Start a fixed-size worker pool; each worker reads jobs and emits results.
	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, client, cfg.retries, jobs, results)
		}(i + 1)
	}

	// Single producer feeds jobs, then closes the channel to signal completion.
	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	// Closer goroutine waits for all workers to finish, then closes results.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Single consumer reads results until close and tallies outcomes.
	total := len(urls)
	var ok, fail, done int
	for r := range results {
		done++
		if total > 0 {
			fmt.Fprintf(os.Stderr, "Progress: %d/%d\r", done, total)
		}
		if r.err != nil {
			fail++
			fmt.Printf("FAIL  %s  err=%v  attempts=%d  %s\n", r.url, r.err, r.attempts, r.duration)
			continue
		}
		ok++
		fmt.Printf("OK    %s  status=%d  attempts=%d  %s\n", r.url, r.status, r.attempts, r.duration)
	}
	if total > 0 {
		fmt.Fprintln(os.Stderr)
	}

	return ok, fail, nil
}
