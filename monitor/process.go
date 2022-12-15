package monitor

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
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

var protocolType6 = []byte{0x60}
var protocolType4 = []byte{0x45}

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

	exclusionList := make([]string, 0)
	path, err := os.Getwd()
	if err == nil {
		file, err := os.Open(path + "/excluded-interfaces")
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				exclusionList = append(exclusionList, scanner.Text())
			}
			_ = file.Close()
		}
	}

	evs := make(chan *eventRaw, 200)
	go InitCounter(evs, len(ifaces))

	var evs2 chan *PortBox
	if analyzePorts {
		evs2 = make(chan *PortBox, 200)
		go InitAnalyzePorts(evs2, len(ifaces))
	}

	for i := range ifaces {
		add := true
		for _, s := range exclusionList {
			if ifaces[i].Name == s {
				add = false
				break
			}
		}
		if add {
			go listen(ifaces[i], evs, evs2)
		}
	}
	log.Println("Start sequence complete")
}

func listen(niface net.Interface, evs chan *eventRaw, evs2 chan *PortBox) {
	tiface := &syscall.SockaddrLinklayer{
		Protocol: htons16(syscall.ETH_P_ALL),
		Ifindex:  niface.Index,
	}
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, htons(syscall.ETH_P_ALL))
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

	protocolType := buf[0:1]
	if bytes.Equal(protocolType, protocolType6) {
		src := make([]byte, 16, 16)
		dst := make([]byte, 16, 16)
		copy(src, buf[8:24])
		copy(dst, buf[24:40])
		evs <- &eventRaw{6, src, dst, uint32(numRead)}
		if analyzePorts {
			data := make([]byte, numRead, numRead)
			copy(data, buf)
			evs2 <- &PortBox{ipVersion: 6, sourceIP: &src, data: &data}
		}
	} else if bytes.Equal(protocolType, protocolType4) {
		src := make([]byte, 4, 4)
		dst := make([]byte, 4, 4)
		copy(src, buf[12:16])
		copy(dst, buf[16:20])
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
