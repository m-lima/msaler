package main

import (
	"fmt"
	"io"
	"os"

	"github.com/manifoldco/promptui"
)

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: msaler [client | COMMAND] [options...]")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "A command-line manager for MSAL clients")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Commands:")
	fmt.Fprintln(writer, "  token   [client] [-v] Generate an oauth token for a client")
	fmt.Fprintln(writer, "  new                   Register a new client")
	fmt.Fprintln(writer, "  delete  [client]      Delete a registered client")
	fmt.Fprintln(writer, "  print   [client]      Print the client information")
	fmt.Fprintln(writer, "  uncache [client]      Print the client information")
	fmt.Fprintln(writer, "  config                Print the path to the configuration file containing the registered clients")
	fmt.Fprintln(writer, "  help                  Print this help message")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Options:")
	fmt.Fprintln(writer, "  client                The client to use. If ommited, a selection menu will be presented")
	fmt.Fprintln(writer, "  -v                    Print extra token fields to stderr")
}

func main() {
	var err error
	if len(os.Args) == 1 {
		err = GetClientToken(os.Args[1:])
	} else {
		modeParam := os.Args[1]

		if "help" == modeParam {
			printUsage(os.Stdout)
		} else if "token" == modeParam {
			err = GetClientToken(os.Args[2:])
		} else if "new" == modeParam {
			err = NewClient(os.Args[2:])
		} else if "delete" == modeParam {
			err = DeleteClient(os.Args[2:])
		} else if "uncache" == modeParam {
			err = UncacheClient(os.Args[2:])
		} else if "print" == modeParam {
			err = PrintClient(os.Args[2:])
		} else if "config" == modeParam {
			var path string
			path, err = ConfigPath()
			if err == nil {
				fmt.Println(path)
			}
		} else {
			err = GetClientToken(os.Args[1:])
			if err != nil {
				err = fmt.Errorf("Error while trying to interpret command as user: %v", err)
			}
		}
	}

	if err != nil && err != promptui.ErrEOF && err != promptui.ErrAbort && err != promptui.ErrInterrupt {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage(os.Stderr)
		os.Exit(-1)
	}
}
