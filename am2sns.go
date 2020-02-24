package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/gorilla/mux"
)

// These variables are defined at compile time
var (
	BuildTime    string = "undefined"
	BuildVersion string = "undefined"
)

// AlertManagerRequest represents the HTTP POST request sent by the Prometheus Alert Manager.
type AlertManagerRequest struct {
	Version           string            `json:"version"`
	GroupKey          string            `json:"groupKey"`
	Status            string            `json:"status"`
	Receiver          string            `json:"receiver"`
	GroupLabels       map[string]string `json:"groupLabels"`
	CommonLabels      map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL       string            `json:"externalURL"`
	Alerts            []struct {
		Status       string            `json:"status"`
		Labels       map[string]string `json:"labels"`
		Annotations  map[string]string `json:"annotations"`
		StartsAt     time.Time         `json:"startsAt"`
		EndsAt       time.Time         `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
	} `json:"alerts"`
}

// Message represents the message formats that is used to send alerts to endpoints.
type Message struct {
	DefaultEndpoint string `json:"default"`
	EmailEndpoint   string `json:"email"`
	SmsEndpoint     string `json:"sms"`
}

func main() {
	initLogger(os.Getenv("LOG_LEVEL"))

	log.Infof("Build time: %s", BuildTime)
	log.Infof("Build version: %s", BuildVersion)

	client := sns.New(session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region:     aws.String(os.Getenv("AWS_SNS_REGION")),
			MaxRetries: aws.Int(3),
		},
	})))

	r := mux.NewRouter()
	r.HandleFunc("/topics/{topicArn}", func(w http.ResponseWriter, r *http.Request) { handleAlert(w, r, client) }).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(handleNotFound)

	s := &http.Server{Handler: r, Addr: ":9876"}
	log.Info("Started listening at ", ":9876")
	log.Fatal(s.ListenAndServe())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Health-check from %s", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Route %s is not handled", r.URL.Path)
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

func handleAlert(w http.ResponseWriter, r *http.Request, client *sns.SNS) {
	// Retrieve target SNS topic ARN from query params
	vars := mux.Vars(r)
	topicArn := vars["topicArn"]
	log.Infof("Handling alert for topic %s", topicArn)

	// Extract Alert Manager request payload
	var amPayload AlertManagerRequest
	err := json.NewDecoder(r.Body).Decode(&amPayload)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Create Email template from Alert Manager data
	emailTpl, err := loadTemplate("data/email.tpl", amPayload)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create SMS template from Alert Manager data
	smsTpl, err := loadTemplate("data/sms.tpl", amPayload)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// SNS needs a specific JSON format to be published
	msg, err := json.Marshal(&Message{
		DefaultEndpoint: "Default endpoint. If you see this, this is probably a configuration issue.",
		EmailEndpoint:   emailTpl,
		SmsEndpoint:     smsTpl,
	})
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Publish alert to SNS topic
	res, err := client.Publish(&sns.PublishInput{
		MessageStructure: aws.String("json"),
		Subject:          aws.String(fmt.Sprintf("[%s:%d] %s", strings.ToUpper(amPayload.Status), len(amPayload.Alerts), amPayload.CommonLabels["alertname"])),
		Message:          aws.String(string(msg)),
		TopicArn:         aws.String(topicArn),
	})
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("Published: %s", res)
	w.WriteHeader(http.StatusAccepted)
}

func initLogger(logLevel string) {
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func loadTemplate(filepath string, data AlertManagerRequest) (string, error) {
	t, err := template.New(filepath).ParseFiles(filepath)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Debugf("Loaded template: %s", tpl.String())
	return tpl.String(), nil
}
