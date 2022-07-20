// Package main defines a log formatter for ingesting mixed text/jSON logs
// outputting something a little less ugly for humans to read.
package main

import (
	"github.com/spf13/cobra"
)

func main() {
	err := cmd.Execute()
	cobra.CheckErr(err)
}
