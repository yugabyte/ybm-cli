// Licensed to Yugabyte, Inc. under one or more contributor license
// agreements. See the NOTICE file distributed with this work for
// additional information regarding copyright ownership. Yugabyte
// licenses this file to you under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func SetFormatter() {
	logrus.SetFormatter(&easy.Formatter{
		LogFormat: "%msg%",
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
