package netcapture

import (
	"encoding/json"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"io/ioutil"
	"os"
)

type Address struct {
	IP string `json:"IP"`
	Port string `json:Port"`
}

func (a *Address) String() string {
	return a.IP + ":" + a.Port
}

type Connection struct {
	Protocol string
	Src Address
	Dst Address
}

func (c *Connection) String() string {
	return c.Protocol + "-" + c.Src.String() + "->" + c.Dst.String()
}

type ConnectionStats struct {
	Count int
}

type NetCaptureConfig {
	Device string `json:"device"`
	Snapshot_len int `json:"snapshot_len"`
	Promiscuous bool `json:"promiscuous"`
}

type NetCapture struct {
	handle *pcap.Handle
	config NetCaptureConfig
	capture_channel chan int
	Connections map[string]*Connection
	Stats map[string]ConnectionStats
}

func (n *NetCapture) addConn(connection *Connection) {
	conn := connection.String()
	if _, ok := n.Connections[conn]; !ok {
		n.Connections[conn] = connection
	}
}

func (n *NetCapture) updateCount(connection *Connection) {
	address := connection.String()
	if as, ok := n.Stats[address]; ok {
		as.Count += 1
		n.Stats[address] = as
	} else {
		as := ConnectionStats{}
		as.Count = 1
		n.Stats[address] = as
	}
}

func (n *NetCapture) processPacket(packet gopacket.Packet) {
	var srcAddr ConnectionStats
	var dstAddr ConnectionStats

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, err := ipLayer.(*layers.IPv4)
		if err != nil {
			//TODO : Log this at some print level
		}

		srcAddr.IP = ip.SrcIP.String()
		dstAddr.IP = ip.DstIP.String()
	}

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, err := tcpLayer.(*layers.TCP)
		if err != nil {
			//TODO : Log this at some print level
		}
		srcAddr.Port = tcp.SrcPort.String()
		dstAddr.Port = tcp.DstPort.String()
	}
	conn := Connection{Protocol:"tcp", Src : srcAddr, Dst : dstAddr}
	n.addConn(conn)
	n.updateCount(conn)
}

func (n *NetCapture) ProcessPackets() {
	packetSource := gopacket.NewPacketSource(n.handle, n.handle.LinkType())
	packet_chan := packetSource.Packets()
	for {
		select {
			case <-n.capture_channel:
				//TODO : Log something here
				return
			case packet := <-packet_chan:
				n.processPacket(packet)
			default:
				continue
		}
	}
}

func (n *NetCapture) StartCapture() {
	go n.ProcessPackets()
}

func (n *NetCapture) StopCapture() {
	n.capture_channel <- 0
	n.handle.Close()
}

func InitializeCapture(config_file string, timeout time.Duration) (*NetCapture, error) {
	file, err := os.Open(config_file)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := NetCaptureConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	handle, err := pcap.OpenLive(config.Device, int32(config.Snapshot_len), config.Promiscuous, timeout)
	if err != nil {
		return nil, err
	}

	n := &NetCapture{handle, config, make(chan int), map[string]NodeStats{}}
	return n, nil
}

