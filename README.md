# `timber` CLI

Command line interface for the [Timber.io](https://timber.io) logging service.

## Installation

Coming soon

## Usage

Use options can be accessed with the `timber-cli help` command:

```
NAME:
   timber - Command line interface for the Timber.io logging service

USAGE:
   timber [global options] command [command options] [arguments...]

COMMANDS:
     tail, t  Live tails logs
     orgs     List organizations that you are a part of
     apps     List applications that you have access to
     views    List saved views that you have access to (currently only console views are displayed)
     api      Make authenticated requests to the Timber API (http://docs.api.timber.io)
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d                Output debug messages [$TIMBER_DEBUG]
   --api-key value, -k value  Your timber.io API key [$TIMBER_API_KEY]
   --host value, -H value     Timber.io host, useful for testing (default: "https://api.timber.io") [$TIMBER_HOST]
   --color-output, -C         Set to force color output even if output is not a color terminal [$TIMBER_COLOR]
   --monochrome-output, -M    Disable color output [$TIMBER_NO_COLOR]
   --help, -h                 show help
   --version, -v              print the version
```