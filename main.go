package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var total = 0

func sendRequests(host string, tls_ bool, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			var conn net.Conn
			var err error
			if tls_ {
				cfg := &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         host,
				}
				conn, err = tls.Dial("tcp", host+":443", cfg)
			} else {
				conn, err = net.Dial("tcp", host+":80")
			}
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			request, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				continue
			}
			if err := request.Write(conn); err != nil {
				continue
			}
			total++
			conn.Close()
		}
	}
}

func main() {
	// Set GOMAXPROCS to utilize all available CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Parameters
	duration := 10 * time.Second
	host := "http://localhost" // change to only selected target where u have permission to do stress testing
	numConnections := 1024
	requestsPerConnection := 1024

	var ttl_ bool
	if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
		ttl_ = true
	} else {
		host = strings.TrimPrefix(host, "http://")
		ttl_ = false
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := 0; k < requestsPerConnection; k++ {
				sendRequests(host, ttl_, done)
			}
		}()
	}
	time.Sleep(duration)
	// Collect and calculate total requests
	elapsed := time.Since(start)
	fmt.Printf("Sent %d requests in %s\n", total, elapsed)
	fmt.Printf("Requests per second: %.2f\n", float64(total)/elapsed.Seconds())
	fmt.Println("Click any key to close...")
	close(done)
	fmt.Scanln()
}
