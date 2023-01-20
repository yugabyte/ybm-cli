package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func SetLogLevel(logLevel string, debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		return
	}
	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Errorln()
			fmt.Fprintf(os.Stderr, "Error Parsing Log level: %s\n", logLevel)
			os.Exit(1)
		}
		logrus.SetLevel(level)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
