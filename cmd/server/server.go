package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	port   int
	rate   int
	rateCh chan struct{}
	delay  time.Duration
}

var cfg config

var (
	latency = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_origin_latency",
		Help: "Origin Latency",
	})
	throughput = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_origin_throughput",
		Help: "Origin Throughput",
	})
)

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	throughput.Inc()
	cfg.rateCh <- struct{}{}
	defer func() {
		<-cfg.rateCh
		throughput.Dec()
		latency.Set(float64(time.Since(start).Milliseconds()) / 1000.0)
	}()
	time.Sleep(cfg.delay)
	_, err := w.Write([]byte("Hello from server\n"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func init() {
	prometheus.MustRegister(latency, throughput)
}

func main() {
	flag.IntVar(&cfg.port, "port", 8080, "port to listen on")
	flag.IntVar(&cfg.rate, "rate", 1, "number of goroutines to run")
	flag.DurationVar(&cfg.delay, "delay", 1*time.Second, "delay per job")
	flag.Parse()
	cfg.rateCh = make(chan struct{}, cfg.rate)
	http.HandleFunc("/", handler)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
