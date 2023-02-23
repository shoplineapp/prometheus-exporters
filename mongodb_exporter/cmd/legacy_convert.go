package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mongodb_performance_exporter/legacy"
)

type LogLine struct {
	date string
}

var (
	logFilePath  string
	destFilePath string
)
var convertCmd = &cobra.Command{
	Short: "convert mongo db log to legacy format before v4.4",
	Run: func(cmd *cobra.Command, args []string) {
		converter := &legacy.LogConverter{}
		if destFilePath == "" {
			converter.ParseFile(logFilePath, nil)
		} else {
			converter.ParseFile(logFilePath, &destFilePath)
		}
	},
}

func Execute() {
	if err := convertCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	convertCmd.Flags().StringVarP(&logFilePath, "file", "f", "", "The input mongo log file")
	convertCmd.MarkFlagRequired("file")
	convertCmd.Flags().StringVarP(&destFilePath, "dest", "d", "", "The output mongo log file")
	convertCmd.MarkFlagRequired("file")
	Execute()
}
