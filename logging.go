package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

type UTCFormatter struct {
	logrus.Formatter
}

func (u UTCFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

// This provides a global logger to be used throughout the Timber package.
// This avoids the need to pass a logger reference around and follows the
// same pattern set in the standard go "log" package. We use zap because
// of it's structured support and performance.
var logger = logrus.New()

// We provide a standard logger in cases where libraries *require* this
// type of logger. This should *not* be used when given a choice. For
// example, the retryablehttp library requires the logger passed to be
// a *log.Logger.
var standardLoggerAlternative = log.New(os.Stdout, "", log.LstdFlags)

func init() {
	// Ensure we're logging in a format that is file friendly.
	textFormatter := &logrus.TextFormatter{DisableColors: true, FullTimestamp: true}
	utcFormatter := UTCFormatter{textFormatter}
	logger.Formatter = utcFormatter
}
