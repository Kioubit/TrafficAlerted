package httpAPI

import (
	"TrafficAlerted/monitor"
	"embed"
	"encoding/json"
	"log"
	"net/http"
)

var moduleName = "mod_httpAPI"

//go:embed dashboard/*
var dashboardContent embed.FS

func init() {
	monitor.RegisterModule(&monitor.Module{
		Name:          moduleName,
		StartComplete: startComplete,
	})
}

func startComplete() {
	http.Handle("/dashboard/", http.FileServer(http.FS(dashboardContent)))
	http.HandleFunc("/active", listActive)
	err := http.ListenAndServe(":8698", nil)
	if err != nil {
		log.Println("["+moduleName+"] Error starting HTTP api server", err.Error())
	}
}

func listActive(w http.ResponseWriter, req *http.Request) {
	marshal, err := json.Marshal(monitor.GetActiveEvents())
	if err != nil {
		w.WriteHeader(500)
		return
	}
	_, _ = w.Write(marshal)
}
