package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}
	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		// log via RPC
		app.logItemViaRPC(w, requestPayload.Log)

		// log via mq
		// app.LogEventViaRabbit(w, requestPayload.Log)

		// log via http
		//app.logItem(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		_ = app.errorJson(w, errors.New("unknown action"))
	}
}

func (app *Config) sendMail(w http.ResponseWriter, mail MailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	request, err := http.NewRequest("POST", "http://mail-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJson(w, errors.New("error calling mail service"))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	if jsonFromService.Error {
		_ = app.errorJson(w, err)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Message sent to " + mail.To

	_ = app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	request, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJson(w, errors.New("error calling logger service"))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	if jsonFromService.Error {
		_ = app.errorJson(w, err)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged!"

	_ = app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to auth microservice
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	fmt.Println("receive authenticate")
	// call the service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	// make sure  we get back the correct status code
	if response.StatusCode == http.StatusUnauthorized {
		_ = app.errorJson(w, err)
		return
	} else if response.StatusCode != http.StatusAccepted {
		_ = app.errorJson(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJson(w, err)
		return
	}
	if jsonFromService.Error {
		_ = app.errorJson(w, err)
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated!"
	payload.Data = jsonFromService.Data

	_ = app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return nil
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}
	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	log.Println("pushToQueue ", string(j))
	return err
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.errorJson(w, err)
		log.Println("Dial failed ", err)
		return
	}
	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}
	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJson(w, err)
		log.Println("call RPC failed ", err)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: result,
	}
	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	log.Println("LogViaGRPC ", requestPayload.Log.Name)

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.errorJson(w, err)
		log.Println("LogViaGRPC conn failed ", requestPayload.Log.Name, err)
		return
	}
	log.Println("LogViaGRPC connected", requestPayload.Log.Name)
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.errorJson(w, err)
		log.Println("LogViaGRPC write failed ", requestPayload.Log.Name, err)
		return
	}
	log.Println("LogViaGRPC write succedd ", requestPayload.Log.Name)
	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJson(w, http.StatusAccepted, payload)

}
