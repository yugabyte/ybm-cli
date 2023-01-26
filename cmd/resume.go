/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// resumeCmd represents the list command
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume resources in YB Managed",
	Long:  "Resume resources in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resumeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// resumeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
