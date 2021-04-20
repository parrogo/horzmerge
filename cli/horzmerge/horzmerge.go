package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/parrogo/horzmerge"
)

// Version of the command
var Version string = "development"

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

var options struct {
	version bool
	out     string
	empty   string
}

func usage(msg string) {
	fmt.Fprintf(os.Stderr, "Wrong usage: %s\n", msg)
	flag.Usage()
	os.Exit(1)
}

func main() {
	flag.BoolVar(&options.version, "v", false, "print version of the command to stdout.")

	flag.StringVar(&options.empty, "e", "", "value of empty cells.")
	flag.StringVar(&options.out, "out", "", "name of output file. Defaults to stdout.")

	flag.Parse()

	if options.version {
		fmt.Println(Version)
		os.Exit(0)
	}
	filenames := flag.Args()
	sources := make([]io.Reader, len(filenames))
	for idx, arg := range filenames {
		f, err := os.Open(arg)
		fatal(err)
		sources[idx] = f
		defer func(f io.Closer) {
			fatal(f.Close())
		}(f)
	}

	opt := horzmerge.Options{
		Empty: options.empty,
	}

	if options.out != "" {
		f, err := os.OpenFile(options.out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
		fatal(err)
		opt.Target = f
		defer func(f io.Closer) {
			fatal(f.Close())
		}(f)
	}

	err := horzmerge.Merge(opt, sources...)
	if err != nil {
		var ierr horzmerge.InputError
		if errors.As(err, &ierr) {
			err = ierr.Convert(filenames)
		}
		fatal(err)
	}
}
