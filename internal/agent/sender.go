package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

// Sender is responsible for sending collected metrics to the server
// via HTTP POST requests.
type Sender struct {
	URL     string
	Storage Storage
	Client  http.Client
}

// Run performs a one-shot send of all stored gauge and counter metrics to the server.
// Counter metrics are reset to zero only after a successful (HTTP 200) response.
func (s *Sender) Run() {

	// Send all gauges
	for metric, value := range s.Storage.GetAllGauges() {
		endpointURL := fmt.Sprintf("%s/update/gauge/%s/%s", s.URL, metric, strconv.FormatFloat(value, 'f', -1, 64))

		request, err := http.NewRequest(http.MethodPost, endpointURL, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain; charset=utf-8")
		resp, err := s.Client.Do(request)

		if err != nil {
			log.Println(err)
			continue
		}
		logRequest(resp, metric, endpointURL)
		drainAndCloseResponse(resp)
	}

	// Send all counters
	for metric, value := range s.Storage.GetAllCounters() {
		endpointURL := fmt.Sprintf("%s/update/counter/%s/%s", s.URL, metric, strconv.FormatInt(value, 10))

		request, err := http.NewRequest(http.MethodPost, endpointURL, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain; charset=utf-8")
		resp, err := s.Client.Do(request)
		if err != nil {
			log.Println(err)
			continue
		}
		resetCounter(resp, metric, s.Storage)
		logRequest(resp, metric, endpointURL)
		drainAndCloseResponse(resp)
	}
}

// resetCounter zeroes the named counter after a successful (HTTP 200) send.
func resetCounter(resp *http.Response, m string, s Storage) {
	// Reset the counter to zero if 200
	if resp.StatusCode == http.StatusOK {
		s.ResetCounter(m)
	}
}

// logRequest logs the outcome of sending a single metric.
func logRequest(resp *http.Response, m string, url string) {
	if resp.StatusCode == http.StatusOK {
		log.Printf("Metric '%s' sent to %s", m, url)
	} else {
		log.Printf("Metric '%s' failed to send to %s, status: %d", m, url, resp.StatusCode)
	}
}

// drainAndCloseResponse drains and closes the response body to allow TCP connection reuse.
func drainAndCloseResponse(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
