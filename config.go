package main

import (
	"TrafficAlerted/monitor"
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func readConfig(path string) monitor.UserConfig {
	file, err := os.Open(path)
	if err != nil {
		showConfigError("Error reading configuration file: " + err.Error())
	}
	scanner := bufio.NewScanner(file)
	read := make(map[string]string)
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if txt == "" || strings.HasPrefix(txt, "#") || !strings.Contains(txt, ":") {
			continue
		}

		line := strings.SplitN(txt, ":", 2)
		if len(line) == 1 {
			continue
		}
		line[1] = strings.SplitN(line[1], "#", 2)[0]
		read[line[0]] = strings.TrimSpace(line[1])
	}
	_ = file.Close()

	// Transfer read keys to config struct
	conf := &monitor.UserConfig{}
	v := reflect.Indirect(reflect.ValueOf(conf))
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		mapRead := read[fieldName]
		if mapRead == "" {
			showConfigError("No value entered for field: " + fieldName)
		}
		switch field.Kind() {
		case reflect.Int:
			input, err := strconv.Atoi(mapRead)
			if err != nil {
				showConfigError(fmt.Sprintf("The value entered for %s is not a number", fieldName))
			}
			if !field.OverflowInt(int64(input)) {
				field.SetInt(int64(input))
			} else {
				showConfigError(fmt.Sprintf("The value entered for %s is too big", fieldName))
			}
		case reflect.Uint32:
			var multiplier uint32 = 1
			switch mapRead[len(mapRead)-1] {
			case 'K':
				multiplier = 1000
			case 'M':
				multiplier = 1000000
			case 'G':
				multiplier = 1000000000
			}
			if multiplier != 1 {
				mapRead = mapRead[:len(mapRead)-1]
			}
			rawNumber, err := strconv.ParseUint(mapRead, 10, 32)
			if err != nil {
				showConfigError(fmt.Sprintf("The value entered for %s is not a number", fieldName))
			}
			input := uint32(rawNumber)
			input = input * multiplier
			if !field.OverflowUint(uint64(input)) {
				field.SetUint(uint64(input))
			} else {
				showConfigError(fmt.Sprintf("The value entered for %s is too big", fieldName))
			}
		case reflect.Bool:
			if !checkIsInputBool(read[fieldName]) {
				showConfigError(fmt.Sprintf("The value entered for %s must be either 'true' or 'false'", fieldName))
			}
			input := mapRead == "true"
			field.SetBool(input)
		case reflect.Slice:
			switch v.Field(i).Type().Elem().Kind() {
			case reflect.String:
				arr := make([]string, 0)
				inText := false
				tmpString := ""
				for i := 0; i < len(mapRead); i++ {
					if mapRead[i] == '"' {
						inText = !inText
						if !inText && len(mapRead)-1 == i {
							arr = append(arr, tmpString)
						}
						continue
					}
					if mapRead[i] == ',' && !inText {
						arr = append(arr, tmpString)
						tmpString = ""
						continue
					}
					tmpString = tmpString + string(mapRead[i])
				}
				field.Set(reflect.ValueOf(arr))
			}
		}
	}
	return *conf
}

func checkIsInputBool(input string) bool {
	if input == "true" || input == "false" {
		return true
	}
	return false
}

func showConfigError(error string) {
	fmt.Println(error)
	os.Exit(1)
}
