package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

type config struct {
	port   int
	rate   int
	rateCh chan struct{}
	delay  time.Duration
}

var cfg config

func handler(w http.ResponseWriter, r *http.Request) {
	cfg.rateCh <- struct{}{}
	defer func() { <-cfg.rateCh }()
	time.Sleep(cfg.delay)
	_, err := w.Write([]byte("Hello from server\n"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	flag.IntVar(&cfg.port, "port", 8080, "port to listen on")
	flag.IntVar(&cfg.rate, "rate", 1, "number of goroutines to run")
	flag.DurationVar(&cfg.delay, "delay", 1*time.Second, "delay per job")
	flag.Parse()
	cfg.rateCh = make(chan struct{}, cfg.rate)
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
