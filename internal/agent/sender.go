package agent

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Sender struct {
	Storage Storage
}

type SenderMonitor struct {
	Interval time.Duration
}

func NewSenderMonitor(interval time.Duration) *SenderMonitor {
	return &SenderMonitor{
		Interval: interval,
	}
}

func (s *SenderMonitor) Run(c *Sender) {
	for {
		client := http.Client{Timeout: 5 * time.Second}

		// Send all gauges
		for metric, value := range c.Storage.GetAllGauges() {
			fullpath := fmt.Sprintf("http://localhost:8080/update/gauge/%s/%s", metric, strconv.FormatFloat(value, 'f', -1, 64))

			request, err := http.NewRequest(http.MethodPost, fullpath, nil)
			request.Header.Add("Content-Type", "text/plain")

			if err != nil {
				fmt.Println(err)
				continue
			}

			res, err := client.Do(request)

			if err != nil {
				fmt.Println(err)
				continue
			}
			res.Body.Close()
			res.Write(os.Stdout)
		}

		// Send all counters
		for metric, value := range c.Storage.GetAllCounters() {
			fullpath := fmt.Sprintf("http://localhost:8080/update/counter/%s/%s", metric, strconv.FormatInt(value, 10))

			request, err := http.NewRequest(http.MethodPost, fullpath, nil)
			request.Header.Add("Content-Type", "text/plain")

			if err != nil {
				fmt.Println(err)
				continue
			}

			res, err := client.Do(request)
			if err != nil {
				fmt.Println(err)
				continue
			}

			res.Body.Close()
			res.Write(os.Stdout)
		}

		time.Sleep(s.Interval)
	}
}
