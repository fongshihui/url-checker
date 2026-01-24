package main

import (
	"fmt"
	"net/http"
	"sync"
)

func run(cfg config, urls []string) (int, int, error) {
	if err := validateConfig(cfg); err != nil {
		return 0, 0, err
	}

	client := &http.Client{Timeout: cfg.timeout}
	jobs := make(chan string)
	results := make(chan result)

	// Start the worker pool and wire up the pipeline.
	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, client, cfg.retries, jobs, results)
		}(i + 1)
	}

	// Feed jobs, then close the channel so workers can exit.
	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	// Close results after all workers have drained the jobs.
	go func() {
		wg.Wait()
		close(results)
	}()

	var ok, fail int
	for r := range results {
		if r.err != nil {
			fail++
			fmt.Printf("FAIL  %s  err=%v  attempts=%d  %s\n", r.url, r.err, r.attempts, r.duration)
			continue
		}
		ok++
		fmt.Printf("OK    %s  status=%d  attempts=%d  %s\n", r.url, r.status, r.attempts, r.duration)
	}

	return ok, fail, nil
}
