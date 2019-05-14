package collector

import (
	"github.com/vaastav/GoNetCapture/netcapture" nc
	"net"
	"net/rpc"
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

func LaunchCollector(config_file string, timeout time.Duration, ip string) error {
	capture, err := nc.InitializeCapture(config_file, timeout)
	if err != nil {
		return err
	}
	addr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		return nil
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil
	}

	collector :=  Collector{capture : capture}
	capture.StartCapture()
	rpc.Register(&collector)
	rpc.Accept(listener)
}
