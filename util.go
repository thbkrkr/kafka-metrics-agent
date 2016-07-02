package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func serveAPI(name string, router *gin.Engine) {
	start := time.Now()

	bindAddr := fmt.Sprintf(":%d", *bindPort)

	s := &http.Server{
		Addr:           bindAddr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	logrus.WithFields(logrus.Fields{
		"in":        time.Since(start),
		"bindAddr":  bindAddr,
		"gitCommit": gitCommit,
		"brokers":   *kafkaBrokers,
		"buildDate": buildDate,
	}).Infof("Start %s", name)

	for {
		s.ListenAndServe()
	}
}

// Base routes

func index(c *gin.Context) {
	c.JSON(200, gin.H{
		"ok":     true,
		"status": 200,
		"name":   "qaas-app",
	})
}

func favicon(c *gin.Context) {
	c.JSON(200, nil)
}

func version(c *gin.Context) {
	c.JSON(200, gin.H{
		"git_commit": gitCommit,
		"build_date": buildDate,
	})
}

// Get calls an URL and unmarshals the JSON response in the value pointed to by obj
func Get(url string, obj interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, obj); err != nil {
		return err
	}

	return nil
}
