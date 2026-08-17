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
		fullpath := fmt.Sprintf("%s/update/gauge/%s/%s", s.URL, metric, strconv.FormatFloat(value, 'f', -1, 64))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain; charset=utf-8")
		res, err := s.Client.Do(request)

		if err != nil {
			log.Println(err)
			continue
		}
		logRequest(res, metric, fullpath)
		finalizeSend(res)
	}

	// Send all counters
	for metric, value := range s.Storage.GetAllCounters() {
		fullpath := fmt.Sprintf("%s/update/counter/%s/%s", s.URL, metric, strconv.FormatInt(value, 10))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain; charset=utf-8")
		res, err := s.Client.Do(request)
		if err != nil {
			log.Println(err)
			continue
		}
		resetCounter(res, metric, s.Storage)
		logRequest(res, metric, fullpath)
		finalizeSend(res)
	}
}

// resetCounter zeroes the named counter after a successful (HTTP 200) send.
func resetCounter(res *http.Response, m string, s Storage) {
	// Reset the counter to zero if 200
	if res.StatusCode == http.StatusOK {
		s.ResetCounter(m)
	}
}

// logRequest logs the outcome of sending a single metric.
func logRequest(res *http.Response, m string, fp string) {
	if res.StatusCode == http.StatusOK {
		log.Printf("Metric '%s' sent to %s", m, fp)
	} else {
		log.Printf("Metric '%s' failed to send to %s, status: %d", m, fp, res.StatusCode)
	}
}

// Drain response body to allow TCP connection reuse (keep-alive).
func finalizeSend(res *http.Response) {
	io.Copy(io.Discard, res.Body)
	res.Body.Close()
}
