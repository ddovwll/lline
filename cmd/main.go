package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/ddovwll/lline"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get working directory: %v\n", err)
		os.Exit(1)
	}

	workers := flag.Int("workers", runtime.GOMAXPROCS(0), "number of workers")
	bufSize := flag.Int("buf-size", 1, "max scanner buffer size in MB")
	rootFlag := flag.String("root", root, "root directory to scan")
	extensionFlag := flag.String("extension", "", "comma-separated file extensions to count lines for")
	excludeFlag := flag.String("exclude", "", "comma-separated substrings of paths to exclude")

	flag.Parse()

	if *extensionFlag == "" {
		fmt.Fprintln(os.Stderr, "flag -extension is required")
		flag.Usage()
		os.Exit(2)
	}

	if *workers <= 0 {
		fmt.Fprintln(os.Stderr, "flag -workers must be greater than 0")
		os.Exit(2)
	}

	if *bufSize <= 0 {
		fmt.Fprintln(os.Stderr, "flag -buf-size must be greater than 0")
		os.Exit(2)
	}

	extensions := strings.Split(*extensionFlag, ",")
	for i := range extensions {
		extensions[i] = strings.TrimSpace(extensions[i])
		if extensions[i] == "" {
			fmt.Fprintln(os.Stderr, "flag -extension must not contain empty values")
			os.Exit(2)
		}
	}

	var exclude []string
	if *excludeFlag != "" {
		exclude = strings.Split(*excludeFlag, ",")
		for i := range exclude {
			exclude[i] = strings.TrimSpace(exclude[i])
			if exclude[i] == "" {
				fmt.Fprintln(os.Stderr, "flag -exclude must not contain empty values")
				os.Exit(2)
			}
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	count, err := lline.CountLines(ctx, *rootFlag, exclude, extensions, *workers, *bufSize)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "count lines: %v\n", err)
		os.Exit(1)
	}

	if errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "shutdown signal received")
	}

	p := message.NewPrinter(language.English)
	fmt.Println(p.Sprintf("%d", count))
}
