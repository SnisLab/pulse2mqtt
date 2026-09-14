package main

import (
	"flag"
	"io"
	"os"
	"testing"

	"pulse2mqtt/include/vars"
)

func TestMainReturnsForHelpFlag(t *testing.T) {
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine
	originalHelp := vars.Help
	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
		vars.Help = originalHelp
	})

	flag.CommandLine = flag.NewFlagSet("pulse2mqtt", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = []string{"pulse2mqtt", "-h"}
	vars.Help = false

	main()

	if !vars.Help {
		t.Fatal("help flag was not parsed")
	}
}
