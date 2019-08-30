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
	host           string
	configFilePath string
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
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(configFilePath)
}

func run(c *cli.Context) error {
	return lib.Run(configFilePath)
}
