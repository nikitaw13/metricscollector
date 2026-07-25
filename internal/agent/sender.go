package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Sender struct {
	URL     string
	Storage Storage
	Client  http.Client
}

// One-shot run
func (s *Sender) Run() {

	// Send all gauges
	for metric, value := range s.Storage.GetAllGauges() {
		fullpath := fmt.Sprintf("%s/update/gauge/%s/%s", s.URL, metric, strconv.FormatFloat(value, 'f', -1, 64))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain")
		res, err := s.Client.Do(request)

		if err != nil {
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric %s have been sent to %s", metric, fullpath)
	}

	// Send all counters
	for metric, value := range s.Storage.GetAllCounters() {
		fullpath := fmt.Sprintf("%s/update/counter/%s/%s", s.URL, metric, strconv.FormatInt(value, 10))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)

		if err != nil {
			log.Println(err)
			continue
		}

		request.Header.Add("Content-Type", "text/plain")
		res, err := s.Client.Do(request)
		if err != nil {
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric '%s' have been sent to %s by sender", metric, fullpath)

		// Reset the counter to zero
		s.Storage.ResetCounter(metric)
	}
}
