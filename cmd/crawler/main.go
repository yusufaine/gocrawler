package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/yusufaine/cs3203-g46-crawler/internal/crawler"
)

func main() {
	// ensures that the data collected so far is exported when the user terminates the program
	// (e.g. ctrl+c)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	config := crawler.SetupConfig()
	config.MustValidate()
	config.PrintConfig()
	time.Sleep(3 * time.Second)

	cr := crawler.New(ctx, config, config.MaxRPS)
	defer cr.GenerateReport(config)

	go func() {
		defer func() {
			cancel()
			close(sig)
		}()

		<-ctx.Done()
		fmt.Println()
		log.Info("stopping crawler", "signal", <-sig)
	}()

	cr.Crawl(ctx, crawler.DefaultLinkExtractor, config.SeedURL, 0)
	log.Info("crawl completed")
}
