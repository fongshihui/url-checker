package main

import (
	"errors"
	"flag"
	"time"
)

// config holds the runtime settings parsed from flags.
type config struct {
	workers int
	timeout time.Duration
	retries int
	inFile  string
}

// parseFlags reads CLI flags and returns the config plus any positional args.
func parseFlags() (config, []string) {
	var (
		workers = flag.Int("workers", 4, "number of concurrent workers")
		timeout = flag.Duration("timeout", 5*time.Second, "request timeout")
		retries = flag.Int("retries", 1, "retry count per URL")
		inFile  = flag.String("in", "", "optional input file with one URL per line")
	)
	flag.Parse()

	return config{
		workers: *workers,
		timeout: *timeout,
		retries: *retries,
		inFile:  *inFile,
	}, flag.Args()
}

// validateConfig enforces basic bounds before starting the worker pool.
func validateConfig(cfg config) error {
	if cfg.workers < 1 {
		return errors.New("workers must be >= 1")
	}
	if cfg.retries < 0 {
		return errors.New("retries must be >= 0")
	}
	return nil
}
