package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nikitaw13/metricscollector/internal/audit"
	"github.com/nikitaw13/metricscollector/internal/handler"
	"github.com/nikitaw13/metricscollector/internal/repository"
	"github.com/nikitaw13/metricscollector/internal/service"
	"go.uber.org/zap"
)

// newExampleServer returns a test HTTP server serving the metric API backed
// by in-memory storage, without a signing key and with no-op auditing.
func newExampleServer() *httptest.Server {
	storage := repository.NewMemStorage()
	svc := service.New(storage, nil, zap.NewNop())
	h := handler.NewMetricsHandler("", svc, zap.NewNop(), audit.New("", "", http.DefaultClient))
	return httptest.NewServer(h.Router())
}

// readBody returns the response body with surrounding whitespace trimmed.
func readBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("read error: %v", err)
	}
	return strings.TrimSpace(string(body))
}

// ExampleMetricsHandler_Router_urlUpdate shows how to update a gauge via the
// plain-text endpoint POST /update/{type}/{metric}/{value}.
func ExampleMetricsHandler_Router_urlUpdate() {
	server := newExampleServer()
	defer server.Close()

	resp, err := http.Post(server.URL+"/update/gauge/temperature/12.5", "text/plain", nil)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d %q\n", resp.StatusCode, readBody(resp))
	// Output: 200 "Metric 'temperature' updated"
}

// ExampleMetricsHandler_Router_urlRead shows how to read a counter via the
// plain-text endpoint GET /value/{type}/{metric}.
func ExampleMetricsHandler_Router_urlRead() {
	server := newExampleServer()
	defer server.Close()

	seedResp, err := http.Post(server.URL+"/update/counter/poll_count/5", "text/plain", nil)
	if err != nil {
		fmt.Println("update request failed:", err)
		return
	}
	seedResp.Body.Close()

	resp, err := http.Get(server.URL + "/value/counter/poll_count")
	if err != nil {
		fmt.Println("read request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d %q\n", resp.StatusCode, readBody(resp))
	// Output: 200 "5"
}

// ExampleMetricsHandler_Router_jsonUpdate shows how to send a single metric
// update as JSON via POST /update/; the response echoes the updated metric.
func ExampleMetricsHandler_Router_jsonUpdate() {
	server := newExampleServer()
	defer server.Close()

	reqBody := `{"id":"cpu","type":"gauge","value":12.5}`
	resp, err := http.Post(server.URL+"/update/", "application/json", strings.NewReader(reqBody))
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d %s %s\n", resp.StatusCode, resp.Header.Get("Content-Type"), readBody(resp))
	// Output: 200 application/json; charset=utf-8 {"id":"cpu","type":"gauge","value":12.5}
}

// ExampleMetricsHandler_Router_jsonUpdates shows how to send a batch of
// metric updates as a JSON array via POST /updates/.
func ExampleMetricsHandler_Router_jsonUpdates() {
	server := newExampleServer()
	defer server.Close()

	reqBody := `[
		{"id":"cpu","type":"gauge","value":12.5},
		{"id":"poll_count","type":"counter","delta":5}
	]`
	resp, err := http.Post(server.URL+"/updates/", "application/json", strings.NewReader(reqBody))
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d %q\n", resp.StatusCode, readBody(resp))
	// Output: 200 ""
}

// ExampleMetricsHandler_Router_jsonRead shows how to ask for the stored
// value of a metric via POST /value/.
func ExampleMetricsHandler_Router_jsonRead() {
	server := newExampleServer()
	defer server.Close()

	seed := `{"id":"cpu","type":"gauge","value":12.5}`
	seedResp, err := http.Post(server.URL+"/update/", "application/json", strings.NewReader(seed))
	if err != nil {
		fmt.Println("seed request failed:", err)
		return
	}
	seedResp.Body.Close()

	reqBody := `{"id":"cpu","type":"gauge"}`
	resp, err := http.Post(server.URL+"/value/", "application/json", strings.NewReader(reqBody))
	if err != nil {
		fmt.Println("read request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d %s\n", resp.StatusCode, readBody(resp))
	// Output: 200 {"id":"cpu","type":"gauge","value":12.5}
}

// ExampleMetricsHandler_Router_list shows how to fetch the HTML page with
// all currently known metrics via GET /.
func ExampleMetricsHandler_Router_list() {
	server := newExampleServer()
	defer server.Close()

	seedResp, err := http.Post(server.URL+"/update/gauge/temperature/12.5", "text/plain", nil)
	if err != nil {
		fmt.Println("update request failed:", err)
		return
	}
	seedResp.Body.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		fmt.Println("list request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%d\n%s\n", resp.StatusCode, readBody(resp))
	// Output: 200
	// <html><body>
	// <h1>List of names and results of all currently known metrics</h1>
	// <p><b>temperature</b>: <code>12.5</code></p>
	// <hr></body></html>
}
