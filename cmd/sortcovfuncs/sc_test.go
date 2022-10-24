// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	start := `
context/context.go:336: init 70.0%
cov.example/p/p.go:15: emptyFnWithBlocks 11.0%
cov.example/p/p.go:24: addStr 100.0%
bytes/buffer.go:404: UnreadByte	0.0%
`
	sr := strings.NewReader(strings.TrimSpace(start))
	var sw strings.Builder
	klines, err := read(sr)
	if err != nil {
		t.Fatalf("read failed with %v", err)
	}
	nlgot := len(klines)
	nlwant := 4
	if nlgot != nlwant {
		t.Fatalf("got %d lines want %d lines", nlgot, nlwant)
	}
	write(&sw, klines)
	got := strings.TrimSpace(sw.String())
	want := strings.TrimSpace(`
cov.example/p/p.go:24: addStr 100.0%
context/context.go:336: init 70.0%
cov.example/p/p.go:15: emptyFnWithBlocks 11.0%
bytes/buffer.go:404: UnreadByte	0.0%
`)
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s\n", got, want)
	}
}

func trygorun(t *testing.T) {
	cmd := exec.Command("go", "run", "testdata/hello.go")
	_, cerr := cmd.CombinedOutput()
	if cerr != nil {
		t.Skipf("go run himom failed, skipping")
	}
}

func TestWithBuild(t *testing.T) {
	trygorun(t)
	if testing.Short() {
		t.Skipf("only during long tests")
	}
	dir := t.TempDir()

	// Generate a coverage profile.
	cp := dir + "/prof.txt"
	cmd := exec.Command("go", "test", "-short", "-coverpkg=all", "-coverprofile", cp, ".")
	out, cerr := cmd.CombinedOutput()
	if cerr != nil {
		t.Logf("%s\n", string(out))
		t.Fatalf("coverage test run failed")
	}

	// Generate a funcs report
	funcs := dir + "/funcs.txt"
	t.Logf("funcs is %s\n", funcs)
	foutf, ferr := os.OpenFile(funcs, os.O_WRONLY|os.O_CREATE, 0666)
	if ferr != nil {
		t.Fatalf("opening %s: %v", funcs, ferr)
	}
	cmd = exec.Command("go", "tool", "cover", "-func", cp)
	cmd.Stdout = foutf
	cerr = cmd.Run()
	if cerr != nil {
		t.Fatalf("go tool cover run failed")
	}
	foutf.Close()

	// A basic run.
	bout := dir + "/basic-output.txt"
	t.Logf("basic output is %s\n", bout)
	os.Args = []string{"sortcovfuncs", "-i", funcs, "-o", bout}
	main()

	// A run with -stripline.
	sout := dir + "/stripped-output.txt"
	t.Logf("basic output is %s\n", sout)
	os.Args[len(os.Args)-1] = sout
	os.Args = append(os.Args, "-stripline")
	main()
	t.Logf("finished main\n")
}
