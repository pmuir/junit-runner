package main

import (
	"bytes"
	json2 "encoding/json"
	"errors"
	"fmt"
	"github.com/clbanning/mxj"
	"github.com/clbanning/mxj/x2j"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var junitXmlRegexp = regexp.MustCompile("TEST-.*\\.xml$")
const url = "http://jenkins-x-reports-elasticsearch-client:9200/tests/junit/"

func main() {
	err := filepath.Walk("target/surefire-reports", func(path string, info os.FileInfo, err error) error {
		if (junitXmlRegexp.MatchString(path)) {
			reader, err := os.Open(path)
			if err != nil {
				return err
			}
			_, json, err := x2j.XmlReaderToJson(reader)
			if err != nil {
				return err
			}
			json, err = munge(json)
			fmt.Printf("Successfully annnotated JUnit result with build info\n")
			if err != nil {
				return err
			}
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))

			req.Header.Set("Content-Type", "application/json")

			if err != nil {
				return err
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if (resp.StatusCode >= 200 && resp.StatusCode < 300 ) {
				fmt.Printf("Sent %s to %s\n", path, url)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				return errors.New(fmt.Sprintf("HTTP status: %s; HTTP Body: %s\n", resp.Status, body))
			}
		}
		return nil
	})
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func munge(json []byte) ([]byte, error) {
	m, err := mxj.NewMapJson(json)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	// Kibana is quite restrictive in the way it accepts JSON, so just rebuild the JSON entirely!
	data := map[string]interface{} {
		"org": os.Getenv("ORG"),
		"appName": os.Getenv("APP_NAME"),
		"version": os.Getenv("VERSION"),
		"errors": m.ValueOrEmptyForPathString("testsuite.-errors"),
		"failures": m.ValueOrEmptyForPathString("testsuite.-failures"),
		"testsuiteName": m.ValueOrEmptyForPathString("testsuite.-name"),
		"skippedTests": m.ValueOrEmptyForPathString("testsuite.-skipped"),
		"tests": m.ValueOrEmptyForPathString("testsuite.-tests"),
		"timestamp": time.Now().Format("2006-01-02T03:04:05Z"),
		// TODO Add the TestCases
	}
	fmt.Printf("%s", data)
	return json2.Marshal(data)
}
