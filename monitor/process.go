package monitor

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"syscall"
)

// Global variables
var (
	noDestination            = false
	cleanupTime              = 1000
	numReadTarget     uint32 = 1000000
	numContactedIPs          = 15
	analyzePorts             = false
	numContactedPorts        = 2
	portCleanupTime          = 1000
)

var etherType6 = []byte{0x86, 0xdd}
var etherType4 = []byte{0x08, 0x00}

type eventRaw struct {
	ipVersion     int
	sourceIP      []byte
	destinationIP []byte
	numRead       uint32
}

type PortBox struct {
	ipVersion int
	sourceIP  *[]byte
	data      *[]byte
}

type UserConfig struct {
	TrafficInterval    int
	ByteLimit          int
	NoDestinations     bool
	NumberContactedIPs int
	AnalyzePorts       bool
	NumContactedPorts  int
	PortInterval       int
}

func Start(conf *UserConfig) {
	noDestination = conf.NoDestinations
	cleanupTime = conf.TrafficInterval
	numReadTarget = uint32(conf.ByteLimit)
	numContactedIPs = conf.NumberContactedIPs
	analyzePorts = conf.AnalyzePorts
	numContactedPorts = conf.NumContactedPorts
	portCleanupTime = conf.PortInterval

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err.Error())
	}
	evs := make(chan *eventRaw, 200)
	go InitCounter(evs, len(ifaces))

	var evs2 chan *PortBox
	if analyzePorts {
		evs2 = make(chan *PortBox, 200)
		go InitAnalyzePorts(evs2, len(ifaces))
	}

	for i := range ifaces {
		go listen(ifaces[i], evs, evs2)
	}
	log.Println("Start sequence complete")
}

func listen(niface net.Interface, evs chan *eventRaw, evs2 chan *PortBox) {
	tiface := &syscall.SockaddrLinklayer{
		Protocol: htons16(syscall.ETH_P_ALL),
		Ifindex:  niface.Index,
	}

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, htons(syscall.ETH_P_ALL))
	if err != nil {
		fmt.Println(err.Error())
	}
	err = syscall.Bind(fd, tiface)
	if err != nil {
		panic(err.Error())
	}

	buf := make([]byte, 5000, 5000)
	for {
		numRead, err := syscall.Read(fd, buf)
		if err != nil {
			panic(err)
		}
		parse(buf, numRead, evs, evs2)
	}

}

func parse(buf []byte, numRead int, evs chan *eventRaw, evs2 chan *PortBox) {
	defer func() {
		if recover() != nil {
			log.Println("[WARN] Recovering from error")
		}
	}()

	etherType := buf[12:14]
	if bytes.Equal(etherType, etherType6) {
		src := make([]byte, 16, 16)
		dst := make([]byte, 16, 16)
		copy(src, buf[22:38])
		copy(dst, buf[38:54])
		evs <- &eventRaw{6, src, dst, uint32(numRead)}
		if analyzePorts {
			data := make([]byte, numRead, numRead)
			copy(data, buf)
			evs2 <- &PortBox{ipVersion: 6, sourceIP: &src, data: &data}
		}
	} else if bytes.Equal(etherType, etherType4) {
		src := make([]byte, 4, 4)
		dst := make([]byte, 4, 4)
		copy(src, buf[26:30])
		copy(dst, buf[30:34])
		evs <- &eventRaw{4, src, dst, uint32(numRead)}
		if analyzePorts {
			data := make([]byte, numRead, numRead)
			copy(data, buf)
			evs2 <- &PortBox{ipVersion: 4, sourceIP: &src, data: &data}
		}
	}
}

func htons16(v uint16) uint16 { return v<<8 | v>>8 }
func htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}
