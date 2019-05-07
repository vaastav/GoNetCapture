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

type AddressStats struct {
	SrcCount int
	DstCount int
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
	Stats map[string]AddressStats
}

func (n *NetCapture) updateSrcDstCount(address string, isSrc bool) {
	if as, ok := n.Stats[address]; ok {
		if isSrc {
			as.SrcCount += 1
		} else {
			as.DstCount += 1
		}
		n.Stats[address] = as
	} else {
		as := AddressStats{}
		if isSrc {
			as.SrcCount = 1
		} else {
			as.DstCount = 1
		}
		n.Stats[address] = as
	}
}

func (n *NetCapture) processPacket(packet gopacket.Packet) {
	var srcAddr AddressStats
	var dstAddr AddressStats

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

	src := srcAddr.String()
	dst := dstAddr.String()
	n.updateSrcDstCount(src, true)
	n.updateSrcDstCount(dst, false)
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

