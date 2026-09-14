package main

import (
	"flag"
	"io"
	"os"
	"testing"
)

func TestMainReturnsForHelpFlag(t *testing.T) {
	originalArgs := os.Args
	originalCommandLine := flag.CommandLine
	t.Cleanup(func() {
		os.Args = originalArgs
		flag.CommandLine = originalCommandLine
	})

	flag.CommandLine = flag.NewFlagSet("pulse2mqtt", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = []string{"pulse2mqtt", "-h"}

	main()
}
