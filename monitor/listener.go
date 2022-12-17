package monitor

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

func listen(niface net.Interface, evs chan *eventRaw, evs2 chan *PortBox) {
	stopWG.Add(1)
	defer stopWG.Done()

	tiface := &syscall.SockaddrLinklayer{
		Protocol: htons16(syscall.ETH_P_ALL),
		Ifindex:  niface.Index,
	}
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, htons(syscall.ETH_P_ALL))
	if err != nil {
		if err.(syscall.Errno) == syscall.EPERM {
			fmt.Println(err.Error())
			fmt.Println("This program needs to be run with the CAP_NET_RAW capability or as root")
			os.Exit(1)
		}
		panic(err.Error())
	}

	err = syscall.Bind(fd, tiface)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		<-stopChan
		_ = syscall.Close(fd)
		stopWG.Done() // syscall.read does not release when the file descriptor is closed
	}()

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

	protocolType := buf[0:1][0]
	var ipVersion int
	var src, dst []byte
	switch parseIPVersion(protocolType) {
	case 6:
		ipVersion = 6
		addresses := make([]byte, 32, 32)
		copy(addresses, buf[8:40])
		src = addresses[0:16]
		dst = addresses[16:32]
	case 4:
		ipVersion = 4
		addresses := make([]byte, 8, 8)
		copy(addresses, buf[12:20])
		src = addresses[0:4]
		dst = addresses[4:8]
	default:
		return
	}

	evs <- &eventRaw{ipVersion, &src, &dst, uint32(numRead)}
	if globalConf.AnalyzePorts {
		data := make([]byte, numRead, numRead)
		copy(data, buf)
		evs2 <- &PortBox{ipVersion: ipVersion, sourceIP: &src, data: &data}
	}

}

func parseIPVersion(v byte) int {
	if uint8(v>>4^byte(6)) == 0 {
		return 6
	} else if uint8(v>>4^byte(4)) == 0 {
		return 4
	}
	return 0
}

func htons16(v uint16) uint16 { return v<<8 | v>>8 }
func htons(v uint16) int {
	return int((v << 8) | (v >> 8))
}
