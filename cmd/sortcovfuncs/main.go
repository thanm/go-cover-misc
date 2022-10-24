// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

var revflag = flag.Bool("r", false, "Reverse sense of sorting")
var striplineflag = flag.Bool("stripline", false, "Strip fn line numbers")
var inflag = flag.String("i", "", "Input file (omit to read from stdin)")
var outflag = flag.String("o", "", "Output file (omit to write to stdout)")

type kline struct {
	perc float32
	line string
}

func dostripline(line string) string {
	fields := strings.Split(line, ":")
	if len(fields) == 3 {
		line = fields[0] + fields[2]
	}
	return line
}

func read(r io.Reader) ([]kline, error) {
	klines := []kline{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("malformed line, not three fields: %s", line)
		}
		perc := float32(0.0)
		n, err := fmt.Sscanf(fields[2], "%f%%", &perc)
		if n != 1 || err != nil {
			return nil, fmt.Errorf("malformed perc %q: %v", fields[2], err)
		}
		if *striplineflag {
			line = dostripline(line)
		}
		k := kline{
			perc: perc,
			line: line,
		}
		klines = append(klines, k)
	}
	return klines, nil
}

func write(w io.Writer, klines []kline) {
	sortfn := func(i, j int) bool {
		return klines[j].perc < klines[i].perc
	}
	if *revflag {
		sortfn = func(i, j int) bool {
			return klines[i].perc < klines[j].perc
		}
	}
	sort.SliceStable(klines, sortfn)
	for i := range klines {
		io.WriteString(w, klines[i].line)
		io.WriteString(w, "\n")
	}
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: sortcovfuncs [flags] -i=<input function report> -o<output file>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func main() {
	flag.Parse()
	var infile *os.File = os.Stdin
	var err error
	closer := func(f *os.File, name string) {
		if err := f.Close(); err != nil {
			fatal("closing %s: %v", name, err)
		}
	}
	if *inflag != "" {
		infile, err = os.Open(*inflag)
		if err != nil {
			fatal("opening %s: %v", *inflag, err)
		}
		defer closer(infile, *inflag)
	}
	var outfile *os.File = os.Stdout
	if *outflag != "" {
		outfile, err = os.OpenFile(*outflag, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fatal("opening %s: %v", *outflag, err)
		}
		defer closer(outfile, *outflag)
	}
	if klines, err := read(infile); err != nil {
		fatal("error reading %s: %v", *inflag, err)
	} else {
		write(outfile, klines)
	}
}
