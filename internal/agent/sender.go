package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Sender struct {
	Storage Storage
}

// One-shot run
func (s *Sender) Run() {

	client := http.Client{Timeout: 5 * time.Second}

	// Send all gauges
	for metric, value := range s.Storage.GetAllGauges() {
		fullpath := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%s", metric, strconv.FormatFloat(value, 'f', -1, 64))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)
		request.Header.Add("Content-Type", "text/plain")

		if err != nil {
			log.Println(err)
			continue
		}

		res, err := client.Do(request)

		if err != nil {
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric %s have been sent to %s", metric, fullpath)
	}

	// Send all counters
	for metric, value := range s.Storage.GetAllCounters() {
		fullpath := fmt.Sprintf("http://localhost:8080/update/counter/%s/%s", metric, strconv.FormatInt(value, 10))

		request, err := http.NewRequest(http.MethodPost, fullpath, nil)
		request.Header.Add("Content-Type", "text/plain")

		if err != nil {
			log.Println(err)
			continue
		}

		res, err := client.Do(request)
		if err != nil {
			log.Println(err)
			continue
		}
		res.Body.Close()
		log.Printf("The metric '%s' have been sent to %s by sender", metric, fullpath)
	}
}
