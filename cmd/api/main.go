package main

import (
	"os"

	"github.com/mlhmz/dockermc-cloud-manager/internal/cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
