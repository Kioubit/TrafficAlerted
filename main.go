package main

import (
	_ "TrafficAlerted/modules"
	"TrafficAlerted/monitor"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("TrafficAlerted")
	conf := &monitor.UserConfig{}

	v := reflect.Indirect(reflect.ValueOf(conf))
	for i := 0; i < v.NumField(); i++ {
		if len(os.Args) != v.NumField()+1 {
			showUsage("invalid number of commandline arguments")
		}
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		switch field.Kind() {
		case reflect.Int:
			input, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				showUsage(fmt.Sprintf("The value entered for %s is not a number", fieldName))
			}
			if !field.OverflowInt(int64(input)) {
				field.SetInt(int64(input))
			} else {
				showUsage(fmt.Sprintf("The value entered for %s is too high", fieldName))
			}
		case reflect.Bool:
			if !checkIsInputBool(os.Args[i+1]) {
				showUsage(fmt.Sprintf("The value entered for %s must be either 'true' or 'false'", fieldName))
			}
			input := os.Args[i+1] == "true"
			field.SetBool(input)
		}
	}

	modules := monitor.GetRegisteredModules()
	for _, m := range modules {
		fmt.Println("Enabled module:", m.Name)
	}
	monitor.ModuleCallback()

	monitor.Start(conf)
	waitForSignal()
}

func showUsage(reason string) {
	if reason != "" {
		fmt.Println("Error:", reason)
	}
	fmt.Println("Usage:", os.Args[0], "<"+strings.Join(getArguments(), `> <`)+`>`)
	fmt.Println("Refer to the documentation for the meaning of those arguments")
	os.Exit(1)
}

func checkIsInputBool(input string) bool {
	if input == "true" || input == "false" {
		return true
	}
	return false
}

func getArguments() []string {
	v := reflect.ValueOf(monitor.UserConfig{})
	args := make([]string, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		args[i] = v.Type().Field(i).Name
	}
	return args
}

func waitForSignal() {
	var sigCh = make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	close(sigCh)
}
