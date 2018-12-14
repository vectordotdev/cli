package main

import (
	"io"
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/aybabtme/rgbterm/rainbow"
	isatty "github.com/mattn/go-isatty"
	"github.com/timberio/timber-cli/api"
)

var version string

// cribbed from fatih/color
var colorize = os.Getenv("TERM") != "dumb" &&
	(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

var client *api.Client

func main() {
	app := cli.NewApp()
	app.Name = "timber"
	app.Usage = "Command line interface for the Timber.io logging service"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "api-key, k",
			Usage:  "Your timber.io API key",
			EnvVar: "TIMBER_API_KEY",
		},
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "Timber.io host, useful for testing",
			Value:  "https://api.timber.io",
			EnvVar: "TIMBER_HOST",
		},
		cli.BoolFlag{
			Name:   "color-output, C",
			Usage:  "Set to force color output even if output is not a color terminal",
			EnvVar: "TIMBER_COLOR",
		},
		cli.BoolFlag{
			Name:   "monochrome-output, M",
			Usage:  "Disable color output",
			EnvVar: "TIMBER_NO_COLOR",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "tail",
			Usage:  "Live tails logs",
			Action: runTail,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "app-id, a",
					Usage:  "The application id(s) to tail. Can be specified multiple times. If empty, will tail all applications.",
					EnvVar: "TIMBER_APP_ID",
				},
				cli.StringFlag{
					Name:   "query, q",
					Usage:  "Query to pass to filter log lines. E.g. level:error",
					EnvVar: "TIMBER_QUERY",
				},
				cli.BoolFlag{
					Name:   "rainbow, r",
					Usage:  "Color your logs with all the colors of the rainbow",
					EnvVar: "TIMBER_RAINBOW",
				},
			},
		},

		{
			Name:  "orgs",
			Usage: "List orgs that you are a part of",
			Action: func(_ *cli.Context) {
				listOrganizations()
			},
			Flags: []cli.Flag{},
		},

		{
			Name:  "applications",
			Usage: "List applications that you have access to",
			Action: func(_ *cli.Context) {
				listApplications()
			},
			Flags: []cli.Flag{},
		},
	}

	app.Before = func(ctx *cli.Context) (err error) {
		if ctx.Bool("color-output") {
			colorize = true
		}
		if ctx.Bool("monochrome-output") {
			colorize = false
		}

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

		client = api.NewClient(host, apiKey)
		client.SetLogger(logger)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		// Exit with 64, EX_USAGE, to indicate a command line usage error
		os.Exit(64)
	}
}

// Entry point for running tail command
func runTail(ctx *cli.Context) error {
	appIds := ctx.StringSlice("app-id")
	if len(appIds) == 0 {
		applications, err := client.ListApplications()
		if err != nil {
			return err
		}
		appIds = make([]string, len(applications))
		log.Printf("found the following applications to tail:")
		for i := range applications {
			appIds[i] = applications[i].Id
			log.Printf("%8s %s", applications[i].Id, applications[i].Name)
		}
	}

	var w io.Writer = os.Stdout
	if ctx.Bool("rainbow") {
		w = rainbow.New(os.Stdout, 252, 255, 43)
		colorize = false
	}

	tail(w, appIds, ctx.String("query"), colorize)

	return nil
}
