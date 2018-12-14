package main

import (
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/aybabtme/rgbterm/rainbow"
	isatty "github.com/mattn/go-isatty"
	"github.com/timberio/timber-cli/api"
)

var version string

var (
	defaultLogFormat = "{{ date }} {{ level }}{{ context.system.ip }} {{ context.system.hostname }} {{ context.http.request_id }} {{ context.user.email }} {{ message }}"
	defaultFacets    = []string{"date", "level", "context.system.hostname"}
)

// cribbed from fatih/color
var colorize = os.Getenv("TERM") != "dumb" &&
	(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

var client *api.Client

func main() {
	app := cli.NewApp()
	app.Name = "timber"
	app.Usage = "Command line interface for the Timber.io logging service"
	app.Version = version

	defaultFacetsSlice := cli.StringSlice(defaultFacets)

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
			Name:  "tail",
			Usage: "Live tails logs",
			Action: func(ctx *cli.Context) error {
				var w io.Writer = os.Stdout
				if ctx.Bool("rainbow") {
					w = rainbow.New(os.Stdout, 252, 255, 43)
					colorize = false // disable colorization so that we don't get conflicting color codes
				}

				var (
					appIds = []string{}
					format = defaultLogFormat
					facets = defaultFacets
					query  = ""
				)

				// pull defaults from view if specified
				if ctx.IsSet("view-id") {
					view, err := client.GetSavedView(ctx.String("view-id"))
					if err != nil {
						return err
					}

					appIds = view.ConsoleSettings.SourceIds
					format = view.ConsoleSettings.LogLineFormat
					if view.ConsoleSettings.Query != nil {
						query = *view.ConsoleSettings.Query
					}
					facets = view.ConsoleSettings.Facets
				}

				if ctx.IsSet("app-id") {
					appIds = ctx.StringSlice("app-id")
				}
				if ctx.IsSet("facet") {
					facets = ctx.StringSlice("facet")
				}
				if ctx.IsSet("log-format") {
					format = ctx.String("log-format")
				}
				if ctx.IsSet("log-format") {
					query = ctx.String("query")
				}

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

				tail(w, appIds, query, format, facets, colorize)

				return nil
			},
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "app-id, a",
					Usage:  "The application id(s) to tail. Can be specified multiple times. If empty, will tail all applications.",
					EnvVar: "TIMBER_APP_ID",
				},
				cli.StringSliceFlag{
					Name:   "facet, F",
					Usage:  "The facets of the logs to highlight. Can be specified multiple times.",
					Value:  &defaultFacetsSlice,
					EnvVar: "TIMBER_APP_ID",
				},
				cli.StringFlag{
					Name:   "view-id, v",
					Usage:  "The view id to tail. If specified, this will set the default app ids, query, and format, but they can be overriden by the appropriate flags.",
					EnvVar: "TIMBER_APP_ID",
				},
				cli.StringFlag{
					Name:   "query, q",
					Usage:  "Query to pass to filter log lines. E.g. level:error.",
					EnvVar: "TIMBER_QUERY",
				},
				// TODO create a new view to get default format if not set
				cli.StringFlag{
					Name:   "log-format, f",
					Usage:  "Template to format log output. Wrap field identifiers with {{ }}. Currently does not output any sort of errors if this cannot be parsed and ignores all non-identifiers.",
					EnvVar: "TIMBER_LOG_FORMAT",
					Value:  defaultLogFormat,
				},
				cli.BoolFlag{
					Name:   "rainbow, r",
					Usage:  "Color your logs with all the colors of the rainbow.",
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

		{
			Name:  "views",
			Usage: "List saved views that you have access to (currently only console views are displayed)",
			Action: func(_ *cli.Context) {
				listSavedViews()
			},
			Flags: []cli.Flag{},
		},

		{
			Name:  "api",
			Usage: "Convenience command for sending requests to the Timber API (http://docs.api.timber.io)",
			Action: func(ctx *cli.Context) error {
				method := strings.ToUpper(ctx.Args().Get(0))
				path := ctx.Args().Get(1)
				return request(method, path, nil)
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
