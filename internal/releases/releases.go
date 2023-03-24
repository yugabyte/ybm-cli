package releases

import (
	"context"

	"github.com/google/go-github/v50/github"
	"github.com/sirupsen/logrus"
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
