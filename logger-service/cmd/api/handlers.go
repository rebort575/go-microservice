package main

import (
	"log"
	"log-service/data"
	"net/http"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JsonPayload
	_ = app.readJson(w, r, &requestPayload)

	log.Println("Write log:", requestPayload.Name, " data:", requestPayload.Data)
	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}
	_ = app.writeJson(w, http.StatusAccepted, resp)
}
