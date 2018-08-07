package cli

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela-sim/version"
	"github.com/urfave/cli"
	"os"
	"path"
)

func Run() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = appUsage
	app.Version = version.Version
	app.Author = author
	app.Email = email

	// Global Flags.
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log, l",
			Value: configuration.DefaultSimLogLevel,
			Usage: "Controls the granularity of the simulator's log traces",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: configuration.DefaultConfigFilePath,
			Usage: "Full path of the simulator configuration file",
		},
	}

	// Before running the user's command.
	app.Before = func(context *cli.Context) error {
		logger := log.New()
		// Set the format of the log text and the place to write
		logger.Level = util.LogLevel(context.String("log"))
		logger.Formatter = util.LogFormatter(true, true)
		logger.SetOutput(os.Stdout)
		util.Init(logger)
		return nil
	}

	app.Commands = commands

	// Run the user's command.
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
