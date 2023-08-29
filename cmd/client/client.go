package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	id         int
	port       int
	targetPort int
	rate       int
}

var cfg config

var (
	successThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("http_success_response_client%d_throughput", cfg.id),
		Help: fmt.Sprintf("Client%d Success Throughput", cfg.id),
	})
	timeoutThroughput = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("http_timeout_response_client%d_throughput", cfg.id),
		Help: fmt.Sprintf("Client%d Timeout Throughput", cfg.id),
	})
)

func addThroughputCounter(isSuccess bool) {
	if isSuccess {
		successThroughput.Inc()
		return
	}
	timeoutThroughput.Inc()
}

func removeThroughputCounter(isSuccess bool) {
	if isSuccess {
		successThroughput.Dec()
		return
	}
	timeoutThroughput.Dec()
}

func init() {
	prometheus.MustRegister(successThroughput, timeoutThroughput)
}

func main() {
	flag.IntVar(&cfg.id, "id", 0, "client id")
	flag.IntVar(&cfg.port, "port", 8081, "port to listen on")
	flag.IntVar(&cfg.targetPort, "targetPort", 8080, "port to send requests to")
	flag.IntVar(&cfg.rate, "rate", 2, "number of requests to send per second")
	flag.Parse()

	successRespCh := make(chan bool, 1)

	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.rate,
			MaxIdleConnsPerHost: cfg.rate,
		},
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil)
		if err != nil {
			fmt.Println(err)
		}
	}()

	var get = func() {
		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d", cfg.targetPort))
		if err != nil {
			var errTimeout *url.Error
			if errors.As(err, &errTimeout) && errTimeout.Timeout() {
				successRespCh <- false
			}
			fmt.Println(err)
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
			successRespCh <- true
		}()
		_, err = io.Copy(os.Stdout, resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Client %d: %s\n", cfg.id, time.Since(start))
	}

	go func() {
		respHistoryQueue := make([]bool, 0, cfg.rate)
		for {
			select {
			case isSuccess := <-successRespCh:
				if len(respHistoryQueue) == cfg.rate {
					lastResp := respHistoryQueue[0]
					removeThroughputCounter(lastResp)
					respHistoryQueue = respHistoryQueue[1:]
				}
				respHistoryQueue = append(respHistoryQueue, isSuccess)
				addThroughputCounter(isSuccess)
			}
		}
	}()

	ticker := time.NewTicker(time.Second / time.Duration(cfg.rate))
	defer ticker.Stop() // Not triggered since the program is killed with Ctrl-C
	for {
		select {
		case <-ticker.C:
			go get()
		}
	}
}
