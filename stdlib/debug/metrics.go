package debug

import (
	"os"
	"log"
	"time"
)

var metricLogger *log.Logger

func ReportMetric(metric string, value interface{}) {
	metricLogger.Println(time.Now().UnixNano(), metric, value)
}

func init() {
	// All metrics should go to stdout!
	metricLogger = log.New(os.Stdout, "", 0)
}