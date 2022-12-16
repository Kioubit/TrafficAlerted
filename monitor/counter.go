package monitor

import (
	"bytes"
	"encoding/hex"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const counterArraySize = 1000

type event struct {
	ipVersion            int
	sourceIP             []byte
	destinationIP        []*[]byte
	destinationIPCounter []uint32
	count                uint32
	notified             bool
	notifiedScanning     bool
}

type eventBox struct {
	sync.Mutex
	event *event
}

var counterArrayPos int32 = 0
var counterArrayMutex sync.RWMutex
var counterArray []*eventBox

var counterArrayLocationMutex sync.RWMutex
var counterArrayLocation map[string]int

func InitCounter(evs chan *eventRaw, ifaceCount int) {
	counterArray = make([]*eventBox, 1000, 1000)
	counterArrayLocation = make(map[string]int)
	for i := 0; i < ifaceCount*4; i++ {
		go worker(evs)
	}
	go cleanup()
}

func cleanup() {
	for {
		time.Sleep(time.Duration(cleanupTime) * time.Second)
		counterArrayLocationMutex.Lock()
		counterArrayMutex.Lock()

		counterArray = make([]*eventBox, counterArraySize, counterArraySize)
		counterArrayLocation = make(map[string]int)
		atomic.StoreInt32(&counterArrayPos, 0)

		counterArrayMutex.Unlock()
		counterArrayLocationMutex.Unlock()
	}
}

func worker(evs chan *eventRaw) {
	for {
		in := <-evs
		if len(evs) > 120 {
			if atomic.CompareAndSwapInt32(&updateDropperRunning, int32(0), int32(1)) {
				go updateDropper(evs)
			}
		}
		assign(in.ipVersion, in.sourceIP, in.destinationIP, in.numRead)
	}
}

func assign(ipVersion int, sourceIP []byte, destinationIP []byte, numRead uint32) {
	counterArrayLocationMutex.RLock()
	loc := counterArrayLocation[hex.EncodeToString(sourceIP)]
	counterArrayLocationMutex.RUnlock()
	if loc == 0 {
		destinationIPArr := make([]*[]byte, 1)
		destinationIPArr[0] = &destinationIP
		destinationIPCounter := make([]uint32, 1)
		destinationIPCounter[0] = numRead

		eventInstance := &event{
			ipVersion:            ipVersion,
			sourceIP:             sourceIP,
			destinationIP:        destinationIPArr,
			destinationIPCounter: destinationIPCounter,
			count:                numRead,
		}
		if numRead > numReadTarget {
			if !eventInstance.notified {
				eventInstance.notified = true
				go notify(eventInstance)
			}
		}
		counterArrayMutex.RLock()
		currentArrPos := int(atomic.AddInt32(&counterArrayPos, 1))
		counterArray[currentArrPos] = &eventBox{event: eventInstance}
		counterArrayLocationMutex.Lock()
		counterArrayLocation[hex.EncodeToString(sourceIP)] = currentArrPos
		counterArrayLocationMutex.Unlock()
		counterArrayMutex.RUnlock()
		if currentArrPos == counterArraySize-2 {
			cleanup()
		}
		return
	}
	counterArrayMutex.RLock()
	box := counterArray[loc]
	box.Lock()
	ev := box.event
	found := false
	for i := range ev.destinationIP {
		if bytes.Equal(*ev.destinationIP[i], destinationIP) {
			ev.destinationIPCounter[i] = Add32(ev.destinationIPCounter[i], numRead)
			found = true
			break
		}
	}
	if !found {
		if !noDestination {
			contactedIPs := len(ev.destinationIPCounter)
			if contactedIPs < 100 {
				ev.destinationIP = append(ev.destinationIP, &destinationIP)
				ev.destinationIPCounter = append(ev.destinationIPCounter, 1)
			}

			if contactedIPs >= numContactedIPs {
				if !ev.notifiedScanning {
					ev.notifiedScanning = true
					notifyScanning(ev)
				}
			}
		}
	}

	ev.count = Add32(ev.count, numRead)

	if ev.count > numReadTarget {
		if !ev.notified {
			ev.notified = true
			go notify(ev)
		}
	}
	box.Unlock()
	counterArrayMutex.RUnlock()

}

func Add32(left, right uint32) uint32 {
	if left > math.MaxInt32-right {
		return left
	}
	return left + right
}

var updateDropperRunning int32 = 0

func updateDropper(updateChannel chan *eventRaw) {
	log.Println("[WARNING] Can't keep up! Dropping some updates")
	for len(updateChannel) > 120 {
		for i := 0; i < 50; i++ {
			<-updateChannel
		}
	}
	log.Println("[INFO] Recovered")
	atomic.StoreInt32(&updateDropperRunning, int32(0))
}
