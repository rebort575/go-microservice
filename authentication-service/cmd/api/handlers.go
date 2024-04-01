package main

import (
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
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}
	log.Println("Authentication email:", requestPayload.Email, ", password:", requestPayload.Password)

	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		log.Println("GetByEmail ", err)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJson(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}
	app.writeJson(w, http.StatusAccepted, payload)
}
