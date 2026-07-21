package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Sender struct {
	Storage Storage
	Client  http.Client
}

// One-shot run
func (s *Sender) Run() {
	baseurl := "http://localhost:8080/update"

	// Send all gauges
	for metric, value := range s.Storage.GetAllGauges() {
		fullpath := fmt.Sprintf("%s/gauge/%s/%s", baseurl, metric, strconv.FormatFloat(value, 'f', -1, 64))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)
		request.Header.Add("Content-Type", "text/plain")

		if err != nil {
			log.Println(err)
			continue
		}

		res, err := s.Client.Do(request)

		if err != nil {
			log.Println(err)
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric %s have been sent to %s", metric, fullpath)
	}

	// Send all counters
	for metric, value := range s.Storage.GetAllCounters() {
		fullpath := fmt.Sprintf("%s/update/counter/%s/%s", baseurl, metric, strconv.FormatInt(value, 10))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)
		request.Header.Add("Content-Type", "text/plain")

		if err != nil {
			log.Println(err)
			log.Println(err)
			continue
		}

		res, err := s.Client.Do(request)
		if err != nil {
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric '%s' have been sent to %s by sender", metric, fullpath)
	}
}
