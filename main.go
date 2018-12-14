package main

import (
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"
)

var version string

func main() {
	app := cli.NewApp()
	app.Name = "timber"
	app.Usage = "Command line interface for the Timber.io logging service"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "api-key",
			Usage:  "Your timber.io API key",
			EnvVar: "TIMBER_API_KEY",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "Timber.io host, useful for testing",
			Value:  "https://api.timber.io",
			EnvVar: "TIMBER_HOST",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "tail",
			Usage:  "Live tails logs",
			Action: runTail,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "app-id",
					Usage:  "The application id(s) to tail. Can be specified multiple times. If empty, will tail all applications.",
					EnvVar: "TIMBER_APP_ID",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		// Exit with 64, EX_USAGE, to indicate a command line usage error
		os.Exit(64)
	}
}

// Entry point for running tail command
func runTail(ctx *cli.Context) error {
	apiKey := ctx.GlobalString("api-key")

	if apiKey == "" {
		message := `Timber API key is not set

We could not locate your Timber API key, please set it via the --api-key flag or by setting the TIMBER_API_KEY env var.`

		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
		return cli.NewExitError(message, 65)
	}

	host := ctx.GlobalString("host")

	if host == "" {
		message := `Timber host is not set

The default is https://api.timber.io, it appears you've overridden this via the --host flag or the TIMBER_HOST env var`

		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
		return cli.NewExitError(message, 65)
	}

	appIds := ctx.StringSlice("app-id")
	if len(appIds) == 0 {
		applications := getApplications(host, apiKey)
		appIds = make([]string, len(applications))
		log.Printf("found the following applications to tail:")
		for i := range applications {
			appIds[i] = applications[i].Id
			log.Printf("%8s %s", applications[i].Id, applications[i].Name)
		}
	}

	tail(host, apiKey, appIds)

	return nil
}
