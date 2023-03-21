package releases

import (
	"context"

	"github.com/google/go-github/v50/github"
)

const (
	org  = "yugabyte"
	repo = "ybm-cli"
)

func GetLatestRelease() (string, error) {
	client := github.NewClient(nil)

	githubRelease, _, err := client.Repositories.GetLatestRelease(context.Background(), org, repo)
	if err != nil {
		return "", err
	}
	releaseVersion := githubRelease.GetTagName()
	return releaseVersion, nil
}
