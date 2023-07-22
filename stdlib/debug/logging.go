package debug

import (
	"os"
	"log"
)

var defaultLogger *log.Logger

func Logger() *log.Logger {
	return defaultLogger
}

func init() {
	defaultLogger = log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds|log.Llongfile)
}
