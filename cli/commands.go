package cli

import "github.com/urfave/cli"

var (
	commands = []cli.Command{
		{
			Name:      "start",
			ShortName: "s",
			Usage:     "Start the engine",
			Category:  "Simulator management",
			Before:    printBanner,
			Action:    start,
		},
	}
)
