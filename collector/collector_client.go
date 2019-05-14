package collector

import (
	ns "github.com/vaastav/GoNetCapture/netcapture"
	"net/rpc"
)

type ColClient struct {
	client *rpc.Client
}

func (c *ColClient) GetStats() (stats ns.Stat, err error) {
	request := 0
	err = c.client.Call("Collector.GetStats", request, &stats)
	return stats, err
}

func (c *ColClient) Close() error {
	if c.client != nil {
		err := c.client.Close()
		return err
	}
	return nil
}

func InitializeClient(serverAddr string) (*ColClient, error) {
	client, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	collector_cl := ColClient{client : client}
	return &collector_cl, nil
}
