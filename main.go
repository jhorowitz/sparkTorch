package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"html/template"

	"flag"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

var (
	siteLocation     string
	bindAddress      string
	sparkId          string
	sparkAccessToken string
)

func init() {
	flag.StringVar(&siteLocation, "site-location", os.Getenv("SITE_LOCATION"), "site entry point")
	flag.StringVar(&bindAddress, "bind-addr", os.Getenv("BIND_ADDR"), "bind address")
	flag.StringVar(&sparkId, "spark-id", os.Getenv("SPARK_ID"), "spark id")
	flag.StringVar(&sparkAccessToken, "spark-access-token", os.Getenv("SPARK_ACCESS_TOKEN"), "spark access token")
	flag.Parse()
	if len(siteLocation) == 0 {
		siteLocation = "./site.html"
	}
}

func main() {
	http.HandleFunc("/", handleMessaging)

	logrus.WithField("address", bindAddress).Debug("HTTP Server started.")
	fatalErr := http.ListenAndServe(bindAddress, nil)
	logrus.WithError(fatalErr).Fatal("Server unexpectedly shut down.")
}

func handleMessaging(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("site.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	} else {
		r.ParseForm()

		if len(r.Form["message"]) == 0 {
			http.Error(w, "Messages must be of at least length 1", http.StatusBadRequest)
			return
		}

		message := r.Form["message"][0]

		logrus.WithField("message", message).Debug("Received message to send.")

		err := SendRequest(message)
		if err != nil {
			logrus.WithError(err).Error("Message send failed.")
			http.Error(w, "Message send failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		t, err := template.ParseFiles("site.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.Execute(w, nil)
	}
}

func SendRequest(message string) error {
	url := fmt.Sprintf("https://api.spark.io/v1/devices/%s/message", sparkId)

	payload := strings.NewReader(fmt.Sprintf("access_token=%s&args=%s", sparkAccessToken, message))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		readAll, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(resp.Status + "\n" + string(readAll))
	}

	return nil
}
