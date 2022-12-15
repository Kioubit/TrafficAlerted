package monitor

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
)

func notify(e *event) {
	defer func() {
		recover()
	}()
	h := convertToHumanEvent(e)
	log.Printf("EVENT=%s SRC=%s TOTALBYTES=%d | DSTINFO=%v\n", h.EventType, h.SourceIP, h.TotalBytes, h.DestinationIP)
}

func notifyScanning(e *event) {
	defer func() {
		recover()
	}()
	h := convertToHumanEvent(e)
	log.Printf("EVENT=%s SRC=%s TOTALBYTES=%d | DSTINFO=%v\n", h.EventType, h.SourceIP, h.TotalBytes, h.DestinationIP)
}

func notifyScanningPort(e PortBox, ports []uint16) {
	defer func() {
		recover()
	}()
	var src string = ""
	if e.ipVersion == 4 {
		src = humanIPv4(*e.sourceIP)
	} else {
		src = humanIPv6(*e.sourceIP)
	}
	log.Printf("EVENT=%s SRC=%s PORTS=%v\n", "SCANNING_PORT", src, ports)
}

func GetActiveEvents() []HumanEvent {
	defer func() {
		recover()
	}()
	var humanEvents = make([]HumanEvent, 0)
	max := atomic.LoadInt32(&counterArrayPos)
	for i := 1; i < int(max); i++ {
		counterArrayMutex.RLock()
		e := counterArray[i].event
		if e.notifiedScanning || e.notified {
			humanEvents = append(humanEvents, convertToHumanEvent(e))
		}
		counterArrayMutex.RUnlock()
	}
	return humanEvents
}

type HumanEvent struct {
	EventType     string
	SourceIP      string
	DestinationIP []HumanDestinationIPInfo
	TotalBytes    uint32
}
type HumanDestinationIPInfo struct {
	IP    string
	Bytes uint32
}

func convertToHumanEvent(e *event) HumanEvent {
	human := &HumanEvent{}
	if e.notifiedScanning {
		human.EventType = "SCANNING"
	} else if e.notified {
		human.EventType = "TRAFFIC"
	} else {
		human.EventType = "TRAFFIC+SCANNING"
	}

	if e.ipVersion == 4 {
		human.SourceIP = humanIPv4(e.sourceIP)
		for i := range e.destinationIP {
			h := HumanDestinationIPInfo{
				IP:    humanIPv4(*e.destinationIP[i]),
				Bytes: e.destinationIPCounter[i],
			}
			human.DestinationIP = append(human.DestinationIP, h)
		}
	} else {
		human.SourceIP = humanIPv6(e.sourceIP)
		for i := range e.destinationIP {
			h := HumanDestinationIPInfo{
				IP:    humanIPv6(*e.destinationIP[i]),
				Bytes: e.destinationIPCounter[i],
			}
			human.DestinationIP = append(human.DestinationIP, h)
		}
	}
	human.TotalBytes = e.count
	return *human
}

func humanIPv4(b []byte) string {
	return net.IPv4(b[0], b[1], b[2], b[3]).String()
}

func humanIPv6(prefix []byte) string {
	return net.IP{prefix[0], prefix[1], prefix[2], prefix[3], prefix[4], prefix[5],
		prefix[6], prefix[7], prefix[8], prefix[9], prefix[10], prefix[11], prefix[12],
		prefix[13], prefix[14], prefix[15]}.String()
}

type Module struct {
	Name          string
	StartComplete func()
}

var (
	moduleList = make([]*Module, 0)
	moduleMu   sync.Mutex
)

func RegisterModule(module *Module) {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	moduleList = append(moduleList, module)
}

func GetRegisteredModules() []*Module {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	return moduleList
}
func ModuleCallback() {
	moduleMu.Lock()
	defer moduleMu.Unlock()
	for _, m := range moduleList {
		if m.StartComplete != nil {
			go m.StartComplete()
		}
	}
}
