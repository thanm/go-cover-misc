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
var inflag = flag.String("i", "", "Input file (omit to read from stdin)")
var outflag = flag.String("o", "", "Output file (omit to write to stdout)")
var mergeflag = flag.Bool("merge", false, "merge together lines with same source position")

type kline struct {
	file  string
	stl   int
	enl   int
	stc   int
	enc   int
	stmts int
	count int
}

func read(r io.Reader, merge bool) ([]kline, string, error) {
	mode := ""
	m := make(map[kline]int)
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
		var stl, enl, stc, enc, stmts, count int
		nv, err := fmt.Sscanf(fields[1], "%d.%d,%d.%d %d %d", &stl, &stc, &enl, &enc, &stmts, &count)
		if err != nil {
			return nil, mode, fmt.Errorf("malformed line (scanf err %v): %s", err, line)
		}
		if nv != 6 {
			return nil, mode, fmt.Errorf("malformed line (scanf %d vals): %s", nv, line)
		}
		k := kline{
			file:  fields[0],
			stl:   stl,
			enl:   enl,
			stc:   stc,
			enc:   enc,
			stmts: stmts,
		}
		if merge {
			v := m[k]
			m[k] = v + count
		} else {
			k.count = count
			klines = append(klines, k)
		}
	}
	if merge {
		for k, v := range m {
			k.count = v
			klines = append(klines, k)
		}
	}
	return klines, mode, nil
}

func write(w io.Writer, mode string, klines []kline) {
	io.WriteString(w, mode)
	io.WriteString(w, "\n")
	lt := func(i, j int) bool {
		if klines[i].file != klines[j].file {
			return klines[i].file < klines[j].file
		}
		if klines[i].stl != klines[j].stl {
			return klines[i].stl < klines[j].stl
		}
		if klines[i].stc != klines[j].stc {
			return klines[i].stc < klines[j].stc
		}
		if klines[i].enl != klines[j].enl {
			return klines[i].enl < klines[j].enl
		}
		return klines[i].enc < klines[j].enc
	}
	sortfn := func(i, j int) bool {
		if *revflag {
			i, j = j, i
		}
		v := lt(i, j)
		return v
	}
	sort.SliceStable(klines, sortfn)
	for _, k := range klines {
		fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n",
			k.file, k.stl, k.stc, k.enl, k.enc, k.stmts, k.count)
	}
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: pnormalize [flags] -i=<input function report> -o<output file>\n")
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
	if klines, mode, err := read(infile, *mergeflag); err != nil {
		fatal("error reading %s: %v", *inflag, err)
	} else {
		write(outfile, mode, klines)
	}
}
