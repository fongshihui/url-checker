package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type apiRequest struct {
	URLs      []string `json:"urls"`
	Workers   *int     `json:"workers,omitempty"`
	TimeoutMS *int     `json:"timeout_ms,omitempty"`
	Retries   *int     `json:"retries,omitempty"`
}

type apiResult struct {
	URL        string `json:"url"`
	Status     int    `json:"status,omitempty"`
	OK         bool   `json:"ok"`
	Error      string `json:"error,omitempty"`
	Attempts   int    `json:"attempts"`
	DurationMS int64  `json:"duration_ms"`
}

type apiResponse struct {
	OK      int         `json:"ok"`
	Fail    int         `json:"fail"`
	Total   int         `json:"total"`
	Results []apiResult `json:"results"`
}

func runServer(cfg config) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req apiRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		if err := ensureNoTrailingJSON(dec); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		urls := normalizeList(req.URLs)
		if len(urls) == 0 {
			http.Error(w, "urls must not be empty", http.StatusBadRequest)
			return
		}

		effective := cfg
		if req.Workers != nil {
			effective.workers = *req.Workers
		}
		if req.TimeoutMS != nil {
			if *req.TimeoutMS <= 0 {
				http.Error(w, "timeout_ms must be > 0", http.StatusBadRequest)
				return
			}
			effective.timeout = time.Duration(*req.TimeoutMS) * time.Millisecond
		}
		if req.Retries != nil {
			effective.retries = *req.Retries
		}
		if err := validateConfig(effective); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results, ok, fail, err := runDetailed(effective, urls)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload := apiResponse{
			OK:      ok,
			Fail:    fail,
			Total:   ok + fail,
			Results: mapResults(results),
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeCORS(w)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "ok\n")
	})

	server := &http.Server{
		Addr:    cfg.addr,
		Handler: mux,
	}

	fmt.Printf("Listening on %s\n", cfg.addr)
	return server.ListenAndServe()
}

func writeCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func ensureNoTrailingJSON(dec *json.Decoder) error {
	var extra interface{}
	err := dec.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("trailing JSON content")
	}
	return err
}

func mapResults(results []result) []apiResult {
	out := make([]apiResult, 0, len(results))
	for _, r := range results {
		entry := apiResult{
			URL:        r.url,
			Status:     r.status,
			OK:         r.err == nil,
			Attempts:   r.attempts,
			DurationMS: r.duration.Milliseconds(),
		}
		if r.err != nil {
			entry.Error = r.err.Error()
		}
		out = append(out, entry)
	}
	return out
}

func runDetailed(cfg config, urls []string) ([]result, int, int, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, 0, 0, err
	}

	client := &http.Client{Timeout: cfg.timeout}
	jobs := make(chan string)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id, client, cfg.retries, jobs, results)
		}(i + 1)
	}

	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var ok, fail int
	out := make([]result, 0, len(urls))
	for r := range results {
		out = append(out, r)
		if r.err != nil {
			fail++
			continue
		}
		ok++
	}

	return out, ok, fail, nil
}
