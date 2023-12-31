package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/yusufaine/gocrawler"
	"github.com/yusufaine/gocrawler/example/tianalyser/internal/tianalyser"
)

func main() {
	// Sends a cancellation signal to the context when ctrl-c is pressed
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	config := tianalyser.SetupConfig()

	// hardcode the seed URL to the TI page
	config.SeedURLs = []string{"https://liquipedia.net/dota2/The_International"}

	// Give the user a few seconds to review the config and a chance to cancel the crawl
	config.PrintConfig()
	time.Sleep(3 * time.Second)

	start := time.Now()

	// New crawler that skips non-OK, non-HTML responses and assumes that every TI page
	// with a country representation links to other TI pages with country representation
	cr := gocrawler.New(ctx,
		&config.Config,
		[]gocrawler.ResponseMatcher{gocrawler.IsHtmlContent},
		tianalyser.TILinkExtractor,
	)

	// Write to file if a panic, cancellation, or completion occurs
	defer func() {
		log.Info("generating TI statisitcs", "file", config.ReportPath)
		tianalyser.Generate(cr, config, time.Since(start))
	}()

	// Ensures that the crawler stops when the context is cancelled (ctrl-c)
	go func() {
		defer func() {
			cancel()
			close(sig)
		}()

		<-ctx.Done()
		fmt.Println()
		log.Info("stopping crawler, press ctrl+c again to force quit", "signal", <-sig)
	}()

	// Start crawling from the seed URL and extract links using the TI link extractor func
	var wg sync.WaitGroup
	for _, seed := range config.SeedURLs {
		wg.Add(1)
		go func(seed string) {
			defer wg.Done()
			cr.Crawl(ctx, 0, seed, "")
		}(seed)
	}
	wg.Wait()
	log.Info("crawl completed")
}
