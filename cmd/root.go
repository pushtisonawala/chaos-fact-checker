package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "chaos-checker",
    Short: "Verify if a Chaos Mesh experiment actually worked",
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(checkCmd)
}