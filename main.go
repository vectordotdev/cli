package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/aybabtme/rgbterm/rainbow"
	"github.com/fatih/color"
	isatty "github.com/mattn/go-isatty"
	"github.com/timberio/cli/api"
	"gopkg.in/urfave/cli.v1"
)

//
// Main variables
//

var (
	apiKey   string
	host     string
	timeZone string
	version  string
)

// cribbed from fatih/color
var colorize = os.Getenv("TERM") != "dumb" &&
	(isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()))

var client *api.Client
var errWriter io.Writer
var infoWriter io.Writer
var successWriter io.Writer
var writer io.Writer

//
// Types
//

type coloredWriter struct {
	Header string
	Color  color.Attribute
	writer io.Writer
}

func (w *coloredWriter) Write(p []byte) (n int, err error) {
	colorFunc := color.New(w.Color).SprintFunc()
	message := strings.TrimRightFunc(string(p), unicode.IsSpace)
	messages := strings.Split(message, "\n")
	nTotal := 0

	if w.Header != "" {
		w.writer.Write([]byte(colorFunc(fmt.Sprint(w.Header))))
	}

	for _, line := range messages {
		line = colorFunc(fmt.Sprint(" ›   ", line, "\n"))

		n, err := w.writer.Write([]byte(line))
		if err != nil {
			return nTotal, err
		}
		nTotal += n
	}

	return nTotal, nil
}

func (w *coloredWriter) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

//
// Errors
//

var ErrNoAPIKey = errors.New("We could not locate your Timber API key, run `timber auth` to login.\n" +
	"Alternatively, you can supply a --api-key flag or set the TIMBER_API_KEY env var.\n" +
	"See https://docs.timber.io/clients/cli#authenticating for more info")

//
// API
//

func init() {
	errWriter = &coloredWriter{Header: " ⚠   Error!\n", Color: color.FgRed, writer: os.Stderr}
	infoWriter = &coloredWriter{Header: " ℹ   Info:\n", Color: color.FgBlue, writer: os.Stdout}
	successWriter = &coloredWriter{Header: " ✓   Success!\n", Color: color.FgGreen, writer: os.Stdout}
	writer = os.Stdout
	cli.ErrWriter = errWriter
}

func main() {
	app := cli.NewApp()
	app.Name = "timber"
	app.Usage = "Command line interface for the Timber.io logging service"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "api-key, K",
			Usage:  "Your timber.io API key",
			EnvVar: "TIMBER_API_KEY",
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
		cli.BoolFlag{
			Name:   "debug, D",
			Usage:  "Output debug messages",
			EnvVar: "TIMBER_DEBUG",
		},
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "Timber.io host, useful for testing",
			Value:  "https://api.timber.io",
			EnvVar: "TIMBER_HOST",
		},
		cli.StringFlag{
			Name:   "time-zone, Z",
			Usage:  "Time zone, such as \"Local\", \"UTC\", or \"America/New_York\"",
			Value:  "Local",
			EnvVar: "TIMBER_TIME_ZONE",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "auth",
			Usage:     "Manage authentication for the Timber CLI",
			ArgsUsage: "[api_key]",
			Flags:     []cli.Flag{},
			Action: func(ctx *cli.Context) error {
				err := setHost(ctx)
				if err != nil {
					return err
				}

				apiKey := ctx.Args().Get(0)

				if apiKey == "" {
					message := "The api_key argument is required: `timber auth [api_key]`\n" +
						"Run `timber help auth` for more details"
					// Exit with 65, EX_DATAERR, to indicate input data was incorrect
					return cli.NewExitError(message, 65)
				}

				organization, err := auth(apiKey)
				if err != nil {
					return err
				}

				message := fmt.Sprint(
					"API key added and set to your active credential\n",
					"Organization ID: ", organization.ID, "\n",
					"Organization Name: ", organization.Name, "\n",
					"Run `timber auth list` to list all credentials.\n",
					"Run `timber auth switch [org_id]` to switch active credentials.\n",
					"Run `timber help auth` for more details",
				)
				successWriter.Write([]byte(message))

				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "list all credentials",
					Action: func(ctx *cli.Context) error {
						err := listCredentials()
						if err != nil {
							return err
						}

						fmt.Println()
						infoWriter.Write([]byte("Run `timber auth switch [org_id]` to switch active credentials\n" +
							"Run `timber auth [api_key]` to add a new credential\n" +
							"Run `timber help auth` for more details"))

						return nil

					},
				},
				{
					Name:      "switch",
					Usage:     "switch active credentials",
					ArgsUsage: "[org_id]",
					Action: func(ctx *cli.Context) error {
						orgID := ctx.Args().Get(0)
						return switchActiveCredentials(orgID)
					},
				},
				{
					Name:      "delete",
					Usage:     "delete a credential",
					ArgsUsage: "[org_id]",
					Action: func(ctx *cli.Context) error {
						orgID := ctx.Args().Get(0)

						if orgID == "" {
							message := "You must supply an org_id: timber auth delete [org_id]"
							// Exit with 65, EX_DATAERR, to indicate input data was incorrect
							return cli.NewExitError(message, 65)
						}

						err := deleteCredential(orgID)
						if err != nil {
							return err
						}

						successWriter.Write([]byte("Credential successfully deleted"))

						return nil

					},
				},
			},
		},

		{
			Name:    "tail",
			Aliases: []string{"t"},
			Usage:   "Live tails logs",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:   "source-id, s",
					Usage:  "The source id(s) to tail. Can be specified multiple times.",
					EnvVar: "TIMBER_SOURCE_ID",
				},
				cli.StringFlag{
					Name:   "view-id, v",
					Usage:  "The view id to tail. If specified, this will set the default app ids, query, and format, but they can be overriden by the appropriate flags.",
					EnvVar: "TIMBER_VIEW_ID",
				},
				cli.StringFlag{
					Name:   "query, q",
					Usage:  "Query to pass to filter log lines. E.g. level:error.",
					EnvVar: "TIMBER_QUERY",
				},
				// TODO create a new view to get default format if not set
				cli.StringFlag{
					Name:   "log-format, f",
					Usage:  "Template to format log output. Must be \"json\" or a custom format. For custom formats, wrap field identifiers with {{ }}. Ex: \"{{ dt }} {{ message }}\". Non-existent fields will be ignored.",
					EnvVar: "TIMBER_LOG_FORMAT",
					Value:  defaultLogFormat,
				},
				cli.BoolFlag{
					Name:   "rainbow, r",
					Usage:  "Color your logs with all the colors of the rainbow.",
					EnvVar: "TIMBER_RAINBOW",
				},
			},
			Action: func(ctx *cli.Context) error {
				err := setGlobalVars(ctx)
				if err != nil {
					return err
				}

				var w io.Writer = os.Stdout
				if ctx.Bool("rainbow") {
					w = rainbow.New(os.Stdout, 252, 255, 43)
					colorize = false // disable colorization so that we don't get conflicting color codes
				}

				var (
					sourceIds = []string{}
					format    = defaultLogFormat
					query     = ""
				)

				// pull defaults from view if specified
				if ctx.IsSet("view-id") {
					view, err := client.GetSavedView(ctx.String("view-id"))
					if err != nil {
						return err
					}

					sourceIds = view.ConsoleSettings.SourceIds
					format = view.ConsoleSettings.LogLineFormat
					if view.ConsoleSettings.Query != nil {
						query = *view.ConsoleSettings.Query
					}
				}

				if ctx.IsSet("source-id") {
					sourceIds = ctx.StringSlice("source-id")
				}

				if len(sourceIds) == 0 {
					message := "You must supply at lease one source ID to tail\n" +
						"1. Run `timber sources` to list all sources\n" +
						"2. Run `timber tail --source-id [source_id]` with the ID of the source you want to tail"
					// Exit with 65, EX_DATAERR, to indicate input data was incorrect
					return cli.NewExitError(message, 65)
				}

				if ctx.IsSet("log-format") {
					format = ctx.String("log-format")
				}

				if ctx.IsSet("query") {
					query = ctx.String("query")
				}

				tail(w, sourceIds, query, format, colorize)

				return nil
			},
		},

		{
			Name:  "sources",
			Usage: "Manage your Timber sources",
			Flags: []cli.Flag{},
			Action: func(ctx *cli.Context) error {
				err := setGlobalVars(ctx)
				if err != nil {
					return err
				}

				err = listSources()
				if err != nil {
					return err
				}

				return nil
			},
		},

		{
			Name:  "sql-queries",
			Usage: "Manage SQL queries",
			Flags: []cli.Flag{},
			Action: func(ctx *cli.Context) error {
				err := setGlobalVars(ctx)
				if err != nil {
					return err
				}

				err = listSQLQueries()
				if err != nil {
					return err
				}

				return nil
			},
			Subcommands: []cli.Command{
				{
					Name:      "execute",
					Usage:     "Execute an SQL query",
					ArgsUsage: "[sql_query]",
					Action: func(ctx *cli.Context) error {
						err := setGlobalVars(ctx)
						if err != nil {
							return err
						}

						query := ctx.Args().Get(0)
						return executeSQLQuery(query)
					},
				},
				{
					Name:      "results",
					Usage:     "Get the results of an SQL query",
					ArgsUsage: "[sql_query_id]",
					Action: func(ctx *cli.Context) error {
						err := setGlobalVars(ctx)
						if err != nil {
							return err
						}

						id := ctx.Args().Get(0)
						return listSQLQueryResults(id)
					},
				},
			},
		},

		{
			Name:  "views",
			Usage: "Manage your saved views (only console views are supported in the CLI)",
			Flags: []cli.Flag{},
			Action: func(ctx *cli.Context) error {
				err := setGlobalVars(ctx)
				if err != nil {
					return err
				}

				return listSavedViews()
			},
		},

		{
			Name:      "api",
			Usage:     "Issue authenticated requests to the Timber API (http://docs.api.timber.io)",
			ArgsUsage: "[method path]",
			Flags:     []cli.Flag{},
			Action: func(ctx *cli.Context) error {
				method := strings.ToUpper(ctx.Args().Get(0))
				path := ctx.Args().Get(1)
				return request(method, path, nil)
			},
		},
	}

	// app.Before = func(ctx *cli.Context) (err error) {
	// 	if ctx.Bool("color-output") {
	// 		colorize = true
	// 	}
	// 	if ctx.Bool("monochrome-output") {
	// 		colorize = false
	// 	}

	// 	apiKey := ctx.GlobalString("api-key")

	// 	if apiKey == "" {
	// 		message := `Timber API key is not set

	// We could not locate your Timber API key, please set it via the --api-key flag or by setting the TIMBER_API_KEY env var.`

	// 		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
	// 		return cli.NewExitError(message, 65)
	// 	}

	// 	host := ctx.GlobalString("host")

	// 	if host == "" {
	// 		message := `Timber host is not set

	// The default is https://api.timber.io, it appears you've overridden this via the --host flag or the TIMBER_HOST env var`

	// 		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
	// 		return cli.NewExitError(message, 65)
	// 	}

	// 	client = api.NewClient(host, apiKey)
	// 	if ctx.Bool("debug") {
	// 		client.SetLogger(logger)
	// 	}

	// 	return nil
	// }

	err := app.Run(os.Args)
	if err != nil {
		errWriter.Write([]byte(err.Error()))
		// Exit with 1, EX_USAGE, to indicate a command line usage error
		os.Exit(1)
	}
}

// This function is called within each `Action` definition. The cli library we are
// using does not offer simple before hooks. The `Before` option overrides a lot of
// default behavior that we do not want to override, so we call this method within
// each action instead.
func setGlobalVars(ctx *cli.Context) error {
	err := setAPIKey(ctx)
	if err != nil {
		return err
	}

	err = setHost(ctx)
	if err != nil {
		return err
	}

	err = setTimeZone(ctx)
	if err != nil {
		return err
	}

	setClient(ctx)

	return nil
}

func setAPIKey(ctx *cli.Context) error {
	apiKey = ctx.GlobalString("api-key")

	if apiKey == "" {
		credential, err := getActiveCredential()
		if err != nil {
			return err
		}
		if credential != nil {
			apiKey = credential.APIKey
		}
	}

	if apiKey == "" {
		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
		return cli.NewExitError(ErrNoAPIKey, 65)
	}

	return nil
}

func setHost(ctx *cli.Context) error {
	host = ctx.GlobalString("host")

	if host == "" {
		message := "Timber host is not set. The default is https://api.timber.io, it appears " +
			"you've overridden this via the --host flag or the TIMBER_HOST env var"

		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
		return cli.NewExitError(message, 65)
	}

	return nil
}

func setTimeZone(ctx *cli.Context) error {
	timeZone = ctx.GlobalString("time-zone")

	if timeZone == "" {
		message := `Time zone is not set. The default is Local, it appears you've overridden "+
			"this via the --time-zone flag or the TIMBER_TIME_ZONE env var`

		// Exit with 65, EX_DATAERR, to indicate input data was incorrect
		return cli.NewExitError(message, 65)
	}

	return nil
}

func setClient(ctx *cli.Context) {
	client = api.NewClient(host, apiKey)
	if ctx.GlobalBool("debug") {
		client.SetLogger(logger)
	}
}
