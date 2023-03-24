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

package releases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
	"github.com/yugabyte/ybm-cli/internal/formatter"
	"golang.org/x/mod/semver"
)

const (
	org  = "yugabyte"
	repo = "ybm-cli"
)

func GetLatestRelease() (string, error) {
	client := github.NewClient(nil)
	// Fetching the latest 10 releases
	opts := &github.ListOptions{
		Page:    1,
		PerPage: 10,
	}
	githubReleases, _, err := client.Repositories.ListReleases(context.Background(), org, repo, opts)
	if err != nil {
		return "", err
	}
	for _, release := range githubReleases {
		logrus.Debugf("Found Release: %s , Prerelease: %v", release.GetTagName(), release.GetPrerelease())
		// Returning the first non-prerelease version
		if !release.GetPrerelease() {
			return release.GetTagName(), nil
		}
	}
	return "", err
}

func PrintUpgradeMessageIfNeeded() {
	// Don't print any error if we are not able to fetch the latest release
	latestVersion, err := GetLatestRelease()
	if err == nil {
		currentVersion := ybmAuthClient.GetVersion()
		logrus.Debugf("Current version: %s, Latest version: %s\n", currentVersion, latestVersion)

		if ShouldUpgrade(currentVersion, latestVersion) {
			message := fmt.Sprintf("A newer version is available. Please upgrade to the latest version %s\n", latestVersion)
			logrus.Println(formatter.Colorize(message, formatter.GREEN_COLOR))
		}
	}
}

func ShouldUpgrade(currentVersion string, latestVersion string) bool {
	// Strip the leading 'v' from the version string
	currentVersion = strings.TrimLeft(currentVersion, "v")
	latestVersion = strings.TrimLeft(latestVersion, "v")
	logrus.Debugf("[Stripped] Current version: %s, Latest version: %s\n", currentVersion, latestVersion)
	return semver.Compare("v"+currentVersion, "v"+latestVersion) == -1
}
