package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
		// TODO

		client := http.Client{Timeout: 5 * time.Second}
		request, err := http.NewRequest(http.MethodPost, "http://localhost:8080", nil)
		request.Header.Add("Content-Type", "text/plain")

		if err != nil {
			log.Fatal(err)
		}

		res, err := client.Do(request)

		if err != nil {
			fmt.Println(err)
		}
		res.Body.Close()
		res.Write(os.Stdout)

		time.Sleep(s.Interval)
	}
}
