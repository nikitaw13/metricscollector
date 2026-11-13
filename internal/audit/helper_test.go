package audit

import (
	"errors"
	"fmt"
	"time"

	"github.com/nikitaw13/metricscollector/internal/model"
)

var (
	testNonExistingPath      = "ABRACADABRA/testfilename.txt"
	testEvent                = []byte("first event")
	testSecondEvent          = []byte("second event")
	testComparingEvent       = []byte("first event\n")
	testComparingDoubleEvent = []byte("first event\nsecond event\n")
)

var (
	testPathOK          = "/ok"
	testPathRedirect    = "/redirect"
	testPathClientError = "/clienterror"
	testPathServerError = "/servererror"
	testPathUnmatched   = "/unmatched"
)

var (
	testUnixTimestamp = time.Now().Unix()
	testMetricName    = "testMetricName"
	testAuditIP       = "127.0.0.1"

	testGaugeValue = float64(1.00)
	testMetric     = model.Metric{
		ID:    testMetricName,
		Type:  model.Gauge,
		Value: &testGaugeValue,
	}

	testJSONEvent = fmt.Appendf([]byte{}, `{"ts":%d,"metrics":[%q],"ip_address":%q}`, testUnixTimestamp, testMetricName, testAuditIP)

	testFilename = "testfilename.txt"
	testURL      = "http://localhost:8181"
)

var (
	errSenderFailed = errors.New("sender failed")
	errCloseFailed  = errors.New("failed to close a subscriber")
)
