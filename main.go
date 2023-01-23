/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/yugabyte/ybm-cli/cmd"
	ybmAuthClient "github.com/yugabyte/ybm-cli/internal/client"
)

var version = "v0.1.0"

func main() {
	ybmAuthClient.SetVersion(version)
	cmd.Execute()
}
