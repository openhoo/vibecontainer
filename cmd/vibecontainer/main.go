package main

import (
	"os"

	"github.com/openhoo/vibecontainer/internal/app"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	a := app.New(version, commit, date)
	if err := a.Execute(); err != nil {
		os.Exit(1)
	}
}
