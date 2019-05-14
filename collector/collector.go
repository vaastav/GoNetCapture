package collector

import (
	"log"
	nc "github.com/vaastav/GoNetCapture/netcapture"
	"net"
	"net/rpc"
	"time"
)

type Collector struct {
	capture *nc.NetCapture
}

func (c* Collector) StopCapture(arg int, ack *bool) error {
	c.capture.StopCapture()
	*ack = true
	return nil
}

func (c* Collector) GetStats(arg int, stats *nc.Stat) error {
	*stats = c.capture.GetStats()
	return nil
}

func LaunchCollector(config_file string, timeout time.Duration, ip string) {
	capture, err := nc.InitializeCapture(config_file, timeout)
	if err != nil {
		log.Fatal(err)
	}
	addr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	collector :=  Collector{capture : capture}
	capture.StartCapture()
	log.Println("Successfully launched collector")
	rpc.Register(&collector)
	rpc.Accept(listener)
}
