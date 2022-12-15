package monitor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

//Analyze tcp + udp

var IPPortMapMu sync.Mutex
var IPPortMap map[string][]uint16

func InitAnalyzePorts(evs chan *PortBox, ifaceCount int) {
	IPPortMap = make(map[string][]uint16)

	for i := 0; i < ifaceCount*4; i++ {
		go analyzeWorker(evs)
	}

	go analyzeCleanup()
}

func analyzeCleanup() {
	for {
		time.Sleep(time.Duration(portCleanupTime) * time.Second)
		IPPortMapMu.Lock()
		IPPortMap = make(map[string][]uint16)
		IPPortMapMu.Unlock()
	}
}

var updateDropperAnalyzeRunning int32 = 0

func analyzeWorker(evs chan *PortBox) {
	for {
		data := <-evs
		if len(evs) > 120 {
			if atomic.CompareAndSwapInt32(&updateDropperAnalyzeRunning, int32(0), int32(1)) {
				go updateDropperAnalyze(evs)
			}
		}
		assignAnalyze(data)
	}
}

func assignAnalyze(data *PortBox) {
	defer func() {
		if recover() != nil {
			log.Println("[WARN] Recovering from error")
		}
	}()
	buf := *data.data
	var protocol []byte
	var protocolStart int
	switch data.ipVersion {
	case 6:
		protocol = buf[20:21]
		protocolStart = 54
		// Note: IPv6 may have more nextHeaders
	case 4:
		protocol = buf[23:24]
		protocolStart = 34
	default:
		return
	}

	buf = buf[protocolStart:]
	var dstport []byte
	if bytes.Equal(protocol, []byte{0x06}) {
		//TCP
		dstport = buf[2:4]
	} else if bytes.Equal(protocol, []byte{0x11}) {
		//UDP
		dstport = buf[2:4]
	} else {
		return
	}
	dstPortHuman := binary.BigEndian.Uint16(dstport)

	IPPortMapMu.Lock()
	arr := IPPortMap[hex.EncodeToString(*data.sourceIP)]
	if arr == nil {
		IPPortMap[hex.EncodeToString(*data.sourceIP)] = []uint16{dstPortHuman}
	} else {
		found := false
		for i := range arr {
			if arr[i] == dstPortHuman {
				found = true
				break
			}
		}
		if !found {
			IPPortMap[hex.EncodeToString(*data.sourceIP)] = append(arr, dstPortHuman)
		}
	}
	if len(arr) > numContactedPorts {
		go notifyScanningPort(*data, arr)
		IPPortMap[hex.EncodeToString(*data.sourceIP)] = []uint16{}
	}
	IPPortMapMu.Unlock()
}

func updateDropperAnalyze(updateChannel chan *PortBox) {
	log.Println("[WARNING] [PortAnalyze] Can't keep up! Dropping some updates")
	for len(updateChannel) > 120 {
		for i := 0; i < 50; i++ {
			<-updateChannel
		}
	}
	log.Println("[INFO] [PortAnalyze] Recovered")
	atomic.StoreInt32(&updateDropperAnalyzeRunning, int32(0))
}
