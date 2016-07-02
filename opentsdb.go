package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

type openTSDBReporter struct {
	url         string
	tokenID     string
	tokenSecret string
	pushChan    chan []Point
}

// Point represents an OpenTSDB time series data point
type Point struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

// StartOpenTSDBReporter starts the OpenTSDB reporter and returns a channel
// to use to send metrics
func StartOpenTSDBReporter(tsURL string, tsToken string) chan []Point {
	if tsURL == "" {
		logrus.Info("OpenTSDB URL is empty, metrics reporting is disabled")
		os.Exit(1)
	}

	splits := strings.Split(tsToken, ":")
	if len(splits) < 2 {
		logrus.Error("Fail to parse OpenTSDB token")
		os.Exit(1)
	}

	tokenID := splits[0]
	tokenSecret := splits[1]

	openTSDBChan := make(chan []Point)

	reporter := openTSDBReporter{
		url:         tsURL,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		pushChan:    openTSDBChan,
	}

	go func() {
		logrus.WithField("URL", tsURL).Info("Start OpenTSDB reporter")
		reporter.run()
	}()

	return openTSDBChan
}

func (r *openTSDBReporter) run() {
	for points := range r.pushChan {
		r.send(points)
	}
}

func (r *openTSDBReporter) send(points []Point) {
	jsonPts, err := json.Marshal(points)
	if err != nil {
		logrus.WithError(err).WithField("points", points).Errorf("Unable to marshal point as JSON")
		return
	}

	err = r.push(bytes.NewBuffer(jsonPts))
	if err != nil {
		logrus.WithError(err).Errorf("Fail to push point to OpenTSDB")
		return
	}

	logrus.WithField("nb", len(points)).Info("Points send to OpenTSDB")
}

func (r *openTSDBReporter) push(data io.Reader) error {
	req, err := http.NewRequest("POST", r.url, data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.tokenID, r.tokenSecret)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(body))
	}
	return nil
}
