/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"github.com/spf13/cobra"
)

// getCmd represents the list command
var ClusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Cluster ",
	Long:  "Cluster command",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	ClusterCmd.AddCommand()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}