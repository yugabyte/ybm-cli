package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func SetFormatter() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors:          viper.GetBool("no-color"),
		DisableLevelTruncation: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			//We don't need the full path to just returning the file
			return "", fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)
		},
	})
}

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
