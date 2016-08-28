package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

var (
	bindAddress      = os.Getenv("BIND_ADDR")
	sparkId          = os.Getenv("SPARK_ID")
	sparkAccessToken = os.Getenv("SPARK_ACCESS_TOKEN")
)

func main() {
	fmt.Println(SendRequest("Hello"))
	return

	logrus.WithField("address", bindAddress).Debug("HTTP Server bind started.")
	fatalErr := http.ListenAndServe(bindAddress, nil)
	logrus.WithError(fatalErr).Fatal("Server unexpectedly shut down.")
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
