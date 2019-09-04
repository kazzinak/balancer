package main

import (
	"fmt"
	"log"
	"os"

	"balancer/lib"

	"github.com/urfave/cli"
)

var (
	version        = "0.1"
	configFilePath string
	isDaemon       = false
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "balancer"
	app.Version = version
	app.Usage = "simple balancer and reverse proxy for dar2019Internship"

	app.Commands = []cli.Command{
		{
			Name:      "run",
			Usage:     "balancer run",
			UsageText: "balancer run [--config-file|-c]",
			Action:    run,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "cf, c",
					Usage:       "path to config file",
					Destination: &configFilePath,
				},
				cli.BoolFlag{
					Name:        "daemon, d",
					Usage:       "daemon flag",
					Destination: &isDaemon,
				},
			},
		},
		{
			Name:      "stop",
			Usage:     "stop command",
			UsageText: "main stop",
			Action:    stop,
		},
		{
			Name:      "reload",
			Usage:     "reload command",
			UsageText: "reload balancer",
			Action:    reload,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(configFilePath)
}

func run(c *cli.Context) error {
	return lib.Run(configFilePath, isDaemon)
}

func stop(c *cli.Context) error {
	return lib.StopServer()
}

func reload(c *cli.Context) error {
	return lib.ReloadServer()
}
