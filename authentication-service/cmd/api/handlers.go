package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authentication(w http.ResponseWriter, r *http.Request) {
	var requestPayload = struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		log.Println("Authentication read json failed")
		_ = app.errorJson(w, err, http.StatusBadRequest)
		return
	}
	log.Println("Authentication email:", requestPayload.Email, ", password:", requestPayload.Password)

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		_ = app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		log.Println("GetByEmail ", err)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		_ = app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	// log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}
	_ = app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return nil
	}
	return nil
}
