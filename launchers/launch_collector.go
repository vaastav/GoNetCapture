package main

import (
	"github.com/vaastav/GoNetCapture/collector"
	"os"
	"log"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage : go run launch_collector.go <path/to/config> <ip:port>")
	}

	config_file := os.Args[1]
	ip := os.Args[2]

	log.Println("Launching collector")
	collector.LaunchCollector(config_file, time.Second * 60, ip)
}
