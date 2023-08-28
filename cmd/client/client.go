package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type config struct {
	id         int
	port       int
	targetPort int
	rate       int
}

var cfg config

func main() {
	flag.IntVar(&cfg.id, "id", 0, "client id")
	flag.IntVar(&cfg.port, "port", 7000, "port to listen on")
	flag.IntVar(&cfg.targetPort, "targetPort", 8080, "port to send requests to")
	flag.IntVar(&cfg.rate, "rate", 2, "number of requests to send per second")
	flag.Parse()
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.rate,
			MaxIdleConnsPerHost: cfg.rate,
		},
	}
	var get = func() {
		start := time.Now()
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d", cfg.targetPort))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		_, _ = io.Copy(os.Stdout, resp.Body)
		fmt.Printf("Client %d: %s\n", cfg.id, time.Since(start))
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop() // Not triggered since the program is killed with Ctrl-C
	for {
		select {
		case <-ticker.C:
			for i := 0; i < cfg.rate; i++ {
				go get()
			}
		}
	}
}
