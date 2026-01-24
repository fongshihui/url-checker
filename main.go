package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cfg, args := parseFlags()

	urls, err := loadURLs(cfg.inFile, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "no URLs provided")
		flag.Usage()
		os.Exit(2)
	}

	ok, fail, err := run(cfg, urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("\nSummary: %d OK, %d FAIL, %d total\n", ok, fail, ok+fail)
}
