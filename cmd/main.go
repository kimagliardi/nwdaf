package main

import (
	"fmt"
	"os"

	"github.com/free5gc/nwdaf/internal/logger"
	"github.com/free5gc/nwdaf/pkg/factory"
	"github.com/free5gc/nwdaf/pkg/service"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var NWDAF = &service.NWDAF{}

func main() {
	app := cli.NewApp()
	app.Name = "nwdaf"
	app.Usage = "5G Network Data Analytics Function (NWDAF)"
	app.Action = action
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
			Value: "config/nwdafcfg.yaml",
		},
		cli.StringFlag{
			Name:  "log, l",
			Usage: "Output log to `FILE`",
		},
		cli.StringFlag{
			Name:  "loglevel",
			Usage: "Set log level (trace|debug|info|warn|error|fatal|panic)",
			Value: "info",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.AppLog.Errorf("NWDAF Run error: %v", err)
		os.Exit(1)
	}
}

func action(c *cli.Context) error {
	// Load configuration
	if err := factory.InitConfigFactory(c.String("config")); err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Initialize logger
	logLevel := c.String("loglevel")
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logger.SetLogLevel(level)
	}

	if logFile := c.String("log"); logFile != "" {
		logger.SetLogFile(logFile)
	}

	logger.AppLog.Infoln("NWDAF version: 1.0.0")
	logger.AppLog.Infoln("Starting NWDAF...")

	// Initialize and start NWDAF service
	NWDAF.Initialize(c)
	NWDAF.Start()

	return nil
}
