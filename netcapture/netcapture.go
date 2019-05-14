package netcapture

import (
	"encoding/json"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"os"
	"sync"
	"time"
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

type Stat struct {
	Stats map[string]ConnectionStats
}

func (c *Connection) String() string {
	return c.Protocol + "-" + c.Src.String() + "->" + c.Dst.String()
}

type ConnectionStats struct {
	Count int
}

type NetCaptureConfig struct {
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
	mux sync.Mutex
}

func (n *NetCapture) addConn(connection Connection) {
	conn := connection.String()
	if _, ok := n.Connections[conn]; !ok {
		n.Connections[conn] = &connection
	}
}

func (n *NetCapture) updateCount(connection Connection) {
	address := connection.String()
	n.mux.Lock()
	if as, ok := n.Stats[address]; ok {
		as.Count += 1
		n.Stats[address] = as
	} else {
		as := ConnectionStats{}
		as.Count = 1
		n.Stats[address] = as
	}
	n.mux.Unlock()
}

func (n *NetCapture) processPacket(packet gopacket.Packet) {
	var srcAddr Address
	var dstAddr Address

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, ok := ipLayer.(*layers.IPv4)
		if !ok {
			log.Println("Packet was not IPv4")
		}

		srcAddr.IP = ip.SrcIP.String()
		dstAddr.IP = ip.DstIP.String()
	}

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, ok := tcpLayer.(*layers.TCP)
		if !ok {
			log.Println("Packet was not TCP")
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
				log.Println("Stopping capture of packets")
				return
			case packet := <-packet_chan:
				n.processPacket(packet)
			default:
				continue
		}
	}
}

func (n *NetCapture) GetStats() Stat {
	n.mux.Lock()
	stat := Stat{Stats : n.Stats}
	n.mux.Unlock()
	return stat
}

func (n *NetCapture) StartCapture() {
	log.Println("Starting Capture")
	go n.ProcessPackets()
}

func (n *NetCapture) StopCapture() {
	n.capture_channel <- 0
	n.handle.Close()
}

func InitializeCapture(config_file string, timeout time.Duration) (*NetCapture, error) {
	log.Println("Initializing capture")
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

	n := &NetCapture{handle : handle, config : config, capture_channel : make(chan int), Connections : map[string]*Connection{}, Stats : map[string]ConnectionStats{}}
	log.Println("Initialization capture complete")
	return n, nil
}

