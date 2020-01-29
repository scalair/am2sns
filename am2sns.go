package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorilla/mux"
)

// AlertManagerPayload represents the HTTP POST request JSON sent by the Prometheus Alert Manager.
type AlertManagerPayload struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []Alert           `json:"alerts"`
}

// Alert represents a Prometheus alert.
type Alert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

// Message represents the message formats that is used to send alerts to endpoints.
type Message struct {
	DefaultEndpoint string `json:"default"`
	EmailEndpoint   string `json:"email"`
	SmsEndpoint     string `json:"sms"`
}

func main() {

	os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	client := sns.New(session.Must(session.NewSession()))

	r := mux.NewRouter()
	r.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) { handleAlert(w, r, client) }).Methods("POST")
	s := &http.Server{
		Handler: r,
		Addr:    ":9876",
	}
	log.Info("Started listening at ", ":9876")
	log.Fatal(s.ListenAndServe())

}

func handleAlert(w http.ResponseWriter, r *http.Request, client *sns.SNS) {
	log.Debug("Entering alert handler")

	topicArn := os.Getenv("AWS_SNS_TOPIC_ARN")

	var payload AlertManagerPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Error("Error when decoding payload: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	payloadJSON, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		log.Error("Error when marshalling payload: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg := &Message{
		DefaultEndpoint: "This is the default endpoint",
		EmailEndpoint:   string(payloadJSON),
		SmsEndpoint:     payload.CommonLabels["alertname"],
	}
	msgJSON, _ := json.Marshal(msg)

	params := &sns.PublishInput{
		MessageStructure: aws.String("json"),
		Subject:          aws.String(fmt.Sprintf("[%s:%d] %s", payload.Status, len(payload.Alerts), payload.CommonLabels["alertname"])),
		Message:          aws.String(string(msgJSON)),
		TopicArn:         aws.String(topicArn),
	}

	res, err := client.Publish(params)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Debug("Alert sent")

	log.Debug(res)
}
