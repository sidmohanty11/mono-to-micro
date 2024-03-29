package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
		Message: "Hit the Broker",
	}
	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
		return
	case "mail":
		app.SendMail(w, requestPayload.Mail)
		return
	case "log":
		app.logItemViaRPC(w, requestPayload.Log)
	default:
		app.errorJSON(w, errors.New("invalid action"), http.StatusBadRequest)
		return
	}
}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "  ")

	req, err := http.NewRequest("POST", "http://auth-service/authenticate", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid creds"), http.StatusUnauthorized)
		return
	} else if resp.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error auth service"), http.StatusInternalServerError)
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(resp.Body).Decode(&jsonFromService)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	if jsonFromService.Error {
		app.errorJSON(w, errors.New(jsonFromService.Message), http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Welcome back"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "  ")

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "http://logger-service/log", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		app.errorJSON(w, errors.New("error logger service"), http.StatusInternalServerError)
		return
	}

	var jsonFromService jsonResponse
	jsonFromService.Error = false
	jsonFromService.Message = "Logging successful"

	app.writeJSON(w, http.StatusAccepted, jsonFromService)
}

func (app *Config) SendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, err := json.MarshalIndent(msg, "", "  ")

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("error mail service"), http.StatusInternalServerError)
		return
	}

	var jsonFromService jsonResponse
	jsonFromService.Error = false
	jsonFromService.Message = "Mail sent to " + msg.To

	app.writeJSON(w, http.StatusAccepted, jsonFromService)
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via RabbitMQ"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}

	err = emitter.Push(string(jsonData), "log.INFO")

	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string

	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = result

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logItemViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload LogPayload

	err := app.readJSON(w, r, &requestPayload)

	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	client := logs.NewLogServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Name,
			Data: requestPayload.Data,
		},
	})

	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged via gRPC"

	app.writeJSON(w, http.StatusAccepted, payload)
}
