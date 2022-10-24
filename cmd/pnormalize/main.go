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
	"strconv"
	"strings"
)

var revflag = flag.Bool("r", false, "Reverse sense of sorting")
var inflag = flag.String("i", "", "Input file (omit to read from stdin)")
var outflag = flag.String("o", "", "Output file (omit to write to stdout)")

type kline struct {
	file string
	st   uint32
	line string
}

func read(r io.Reader) ([]kline, string, error) {
	mode := ""
	klines := []kline{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			return nil, mode, fmt.Errorf("malformed line, not two fields: %s", line)
		}
		if fields[0] == "mode" {
			mode = line
			continue
		}
		fields2 := strings.Split(fields[1], ".")
		if len(fields2) != 3 {
			return nil, mode, fmt.Errorf("malformed line, bad lines clause: %s", line)
		}
		sl, err := strconv.Atoi(fields2[0])
		if err != nil {
			return nil, mode, fmt.Errorf("malformed starting line number: %s", line)
		}
		k := kline{
			file: fields[0],
			st:   uint32(sl),
			line: line,
		}
		klines = append(klines, k)
	}
	return klines, mode, nil
}

func write(w io.Writer, mode string, klines []kline) {
	io.WriteString(w, mode)
	io.WriteString(w, "\n")
	sortfn := func(i, j int) bool {
		if klines[i].file != klines[j].file {
			return klines[i].file < klines[j].file
		}
		return klines[i].st < klines[j].st
	}
	if *revflag {
		sortfn = func(i, j int) bool {
			return sortfn(j, i)
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
	if klines, mode, err := read(infile); err != nil {
		fatal("error reading %s: %v", *inflag, err)
	} else {
		write(outfile, mode, klines)
	}
}
