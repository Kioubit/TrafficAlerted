package monitor

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

var globalConf UserConfig

type UserConfig struct {
	TrafficInterval     int
	ByteLimit           uint32
	NoDestinations      bool
	ContactedIPsAnalyze bool
	NumberContactedIPs  int
	AnalyzePorts        bool
	NumContactedPorts   int
	PortInterval        int
	ChartMinBytes       uint32
	ExcludedInterfaces  []string
}

var stopWG sync.WaitGroup
var stopChan = make(chan interface{})

type eventRaw struct {
	ipVersion     int
	sourceIP      *[]byte
	destinationIP *[]byte
	numRead       uint32
}

type PortBox struct {
	ipVersion int
	sourceIP  *[]byte
	data      *[]byte
}

func Start(suppliedConf UserConfig) {
	globalConf = suppliedConf
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err.Error())
	}

	evs := make(chan *eventRaw, 200)
	go InitCounter(evs, len(ifaces))

	var evs2 chan *PortBox
	if globalConf.AnalyzePorts {
		evs2 = make(chan *PortBox, 200)
		go InitAnalyzePorts(evs2, len(ifaces))
	}

	for i := range ifaces {
		add := true
		for _, s := range globalConf.ExcludedInterfaces {
			if ifaces[i].Name == s {
				add = false
				break
			}
		}
		if add {
			go listen(ifaces[i], evs, evs2)
		}
	}
	log.Println("Started")
}

func Stop() {
	fmt.Println("Shutting down instance..")
	close(stopChan)
	if wgWaitTimout(&stopWG, 10*time.Second) {
		fmt.Println("Done")
	} else {
		fmt.Println("Error shutting down instance")
	}
}

func wgWaitTimout(wg *sync.WaitGroup, timeout time.Duration) bool {
	t := make(chan struct{})
	go func() {
		defer close(t)
		wg.Wait()
	}()
	select {
	case <-t:
		return true
	case <-time.After(timeout):
		return false
	}
}
