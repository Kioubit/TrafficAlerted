package main

import (
	_ "TrafficAlerted/modules"
	"TrafficAlerted/monitor"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var Version = ""

func main() {
	if len(os.Args) != 2 {
		showUsage()
	}
	conf := readConfig(os.Args[1])

	fmt.Println("TrafficAlerted", Version)
	modules := monitor.GetRegisteredModules()
	for _, m := range modules {
		fmt.Println("Enabled module:", m.Name)
	}
	monitor.ModuleCallback()

	monitor.Start(conf)
	waitForSignal()
	monitor.Stop()
}

func showUsage() {
	fmt.Println("Usage:", os.Args[0], "<configuration file path>")
	os.Exit(1)
}

func waitForSignal() {
	var sigCh = make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	close(sigCh)
}
