/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// pauseCmd represents the list command
var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause resources in YB Managed",
	Long:  "Pause resources in YB Managed",
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("pause called")
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pauseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pauseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
