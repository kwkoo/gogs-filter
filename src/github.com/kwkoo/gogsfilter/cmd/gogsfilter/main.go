package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/kwkoo/configparser"
	"github.com/kwkoo/gogsfilter"
)

func main() {
	config := struct {
		ListenPort int    `json:"port" default:"8080" usage:"Web server port"`
		Rulesjson  string `json:"rulesjson" usage:"Mapping rules in JSON format; rules are expected to be a slice of objects with a ref key and a target key; if target contains {{}} it is assumed to be a template"`
	}{}
	if err := configparser.Parse(&config); err != nil {
		log.Printf("error parsing configuration: %v", err)
		os.Exit(1)
	}

	fc := gogsfilter.InitFilterConfig(config.Rulesjson)

	// Setup signal handling.
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)

	var wg sync.WaitGroup
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ListenPort),
		Handler: fc,
	}
	go func() {
		log.Printf("listening on port %v", config.ListenPort)
		wg.Add(1)
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Print("web server graceful shutdown")
				return
			}
			log.Fatal(err)
		}
	}()

	// Wait for SIGINT
	<-shutdown
	log.Print("interrupt signal received, initiating web server shutdown...")
	signal.Reset(os.Interrupt)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)

	wg.Wait()
	log.Print("Shutdown successful")
}
