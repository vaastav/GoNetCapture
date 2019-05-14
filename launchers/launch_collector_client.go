package main

import (
	"github.com/vaastav/GoNetCapture/collector"
	"os"
	"log"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage : go run launch_collector_client.go <collector_ip:collector_port>")
	}

	ip := os.Args[1]

	log.Println("Launching collector client")
	client, err := collector.InitializeClient(ip)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(60 * time.Second)
	stats, err := client.GetStats()
	if err != nil {
		log.Fatal(err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(stats)
}
