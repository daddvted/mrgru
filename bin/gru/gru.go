package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"

	"github.com/ski2per/gru/gru"
	"github.com/ski2per/gru/gru/web"
)

var g = gru.Gru{}

func init() {
	// Print version
	fmt.Printf("\nGru: %s\n\n", gru.Version)

	// Init configuration from environmental variables.
	if err := env.Parse(&g); err != nil {
		log.Errorf("%+v\n", err)
	}
}

func main() {
	// Do bootstraping checks
	bootstrap()

	log.Infof("hello %s\n", g.Name)
	sigs := make(chan os.Signal, 1)
	// Notify signal like "Ctrl+C"
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		shutDown(ctx, sigs, cancel)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		web.RunHTTPServer(ctx, g.Mode)
		wg.Done()
	}()

	wg.Wait()
	log.Info("Exited...")
	os.Exit(0)
}

func bootstrap() {
	// Init Logrus, default log level is XXX.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.00000",
	})
	logLvl, err := log.ParseLevel(g.LogLevel)
	if err != nil {
		// Set level to INFO when error happened.
		logLvl = log.InfoLevel
	}
	log.SetLevel(logLvl)

}

func shutDown(ctx context.Context, sigs chan os.Signal, cancel context.CancelFunc) {
	// Waiting for context do be done or TERM signal(ctrl+C)
	select {
	case <-ctx.Done():
		log.Info("Shutting down...")
	case <-sigs:
		// Call cancel on the context to close everything down.
		cancel()
		log.Info("Shutdown with cancel signal...")
	}

	// Unregister to get default OS nuke behaviour in case we don't exit cleanly
	signal.Stop(sigs)
}
