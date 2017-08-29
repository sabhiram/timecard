package main

////////////////////////////////////////////////////////////////////////////////

import (
	"flag"
	"log"
	"os"
)

////////////////////////////////////////////////////////////////////////////////

const (
	timecardFile = ".timecard"
	version      = "0.0.1"
	usage        = `usage: timecard [--version] [--help] <command> [<args>]

Valid Timecard commands include:
    init        Create an empty timecard or re-initialize an existing one
    start       Start or re-start the timecard for the current commit
    checkpoint  Create a checkpoint within a given interval
    end         End a timestamp with a given tag (usually a commit hash)
`
)

////////////////////////////////////////////////////////////////////////////////

var (
	CLI = struct {
		version bool // print application version
		help    bool // print application usages
	}{}
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	args := flag.Args()

	if CLI.version {
		log.Printf("%s\n", version)
	} else if CLI.help || len(args) == 0 {
		log.Printf("%s\n", usage)
	} else {
		log.Printf("Got args: %#v\n", args)
	}
}

////////////////////////////////////////////////////////////////////////////////

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	flag.BoolVar(&CLI.version, "version", false, "print the application version")
	flag.BoolVar(&CLI.version, "v", false, "print the application version (short)")
	flag.BoolVar(&CLI.help, "help", false, "print the application help")
	flag.BoolVar(&CLI.help, "h", false, "print the application help (short)")
	flag.Parse()
}

////////////////////////////////////////////////////////////////////////////////
