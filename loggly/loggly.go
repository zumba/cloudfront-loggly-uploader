package loggly

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/zumba/cloudfront-loggly-uploader/logging"
)

// Loggly has a 5MB limit on posts via bulk API
const logglyBufferLimit = 5 * 1024 * 1024

// Loggly bulk API endpoint URL
const logglyAPIBulkEndpoint = "https://logs-01.loggly.com/bulk"

// Number of times to try HTTP POST in case there's an error
const logglyHTTPRetries = 2

// Sleep time for exponential backoff between retries
const logglyHTTPSleepSecs = 2

// Client stores JSON log entries in a buffer to be sent at a later time
type Client struct {
	apiKey string
	tag    string
	buffer bytes.Buffer
}

// NewClient creates a new Client object
func NewClient(apiKey, tag string, debug bool) *Client {
	client := Client{}
	client.apiKey = apiKey
	client.tag = tag

	return &client
}

// AddToBuffer writes a log to Loggly buffer
func (lc *Client) AddToBuffer(logentry string) error {
	logging.DebugLogger.Printf("Loggly buffer length: %d\n", lc.buffer.Len())

	if lc.buffer.Len()+len(logentry) < logglyBufferLimit {
		lc.buffer.Write([]byte(logentry))
	} else {
		log.Printf("Loggly buffer limit would be exceeded with write: %d\n", len(logentry))
		lc.SendBuffer()
		lc.buffer.Write([]byte(logentry))
	}

	return nil
}

// SendBuffer sends logs to Loggly using HTTP bulk API and flushes
// byte buffer so we can accept more log entries
func (lc *Client) SendBuffer() error {
	if lc.buffer.Len() == 0 {
		log.Printf("Loggly buffer length is 0, not flushing\n")
		return nil
	}

	log.Printf("Sending buffer to Loggly: size=%d\n", lc.buffer.Len())

	// Try making HTTP request twice before giving up
	var err error
	for i := 0; i < logglyHTTPRetries; i++ {
		err = lc.makeHTTPPostRequest()

		if err == nil {
			break
		} else {
			log.Printf("Unable to send buffer via POST: try=%d err=%s\n", i, err)
			time.Sleep(time.Duration(logglyHTTPRetries) * time.Duration(i) * time.Second)
			continue
		}
	}

	lc.buffer.Reset()
	return err
}

// makeHTTPPostRequest sends Loggly buffer via HTTP POST request
func (lc *Client) makeHTTPPostRequest() error {
	url := fmt.Sprintf("%s/%s/tag/%s", logglyAPIBulkEndpoint, lc.apiKey, lc.tag)

	reader := bytes.NewReader(lc.buffer.Bytes())
	resp, err := http.Post(url, "text/plain", reader)
	if err != nil {
		log.Printf("Error sending buffer to Loggly via HTTP POST: %s", err)
		return err
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)

	logging.DebugLogger.Printf("Loggly POST response: %s", body)
	return nil
}
