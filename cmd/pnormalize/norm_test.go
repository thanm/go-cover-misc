// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	start := `
runtime/coverage/apis.go:183.185,18.3 1 0
runtime/coverage/apis.go:156.156,2.38 1 0
runtime/coverage/emit.go:565.565,4.15 1 0
runtime/coverage/emit.go:496.497,39.32 1 0
`
	sr := strings.NewReader(strings.TrimSpace(start))
	var sw strings.Builder
	klines, mode, err := read(sr, false)
	if err != nil {
		t.Fatalf("read failed with %v", err)
	}
	nlgot := len(klines)
	nlwant := 4
	if nlgot != nlwant {
		t.Fatalf("got %d lines want %d lines", nlgot, nlwant)
	}
	write(&sw, mode, klines)
	got := strings.TrimSpace(sw.String())
	want := strings.TrimSpace(`
runtime/coverage/apis.go:156.156,2.38 1 0
runtime/coverage/apis.go:183.185,18.3 1 0
runtime/coverage/emit.go:496.497,39.32 1 0
runtime/coverage/emit.go:565.565,4.15 1 0
`)
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s\n", got, want)
	}
}

func TestMerge(t *testing.T) {
	start := `
runtime/coverage/apis.go:183.185,18.3 1 0
runtime/coverage/apis.go:183.185,18.3 1 1
runtime/coverage/apis.go:156.156,2.38 1 0
runtime/coverage/emit.go:565.565,4.15 1 0
runtime/coverage/emit.go:565.565,4.15 1 99
runtime/coverage/emit.go:496.497,39.32 1 0
`
	sr := strings.NewReader(strings.TrimSpace(start))
	var sw strings.Builder
	klines, mode, err := read(sr, true)
	if err != nil {
		t.Fatalf("read failed with %v", err)
	}
	nlgot := len(klines)
	nlwant := 4
	if nlgot != nlwant {
		t.Logf("klines: %+v\n", klines)
		t.Fatalf("got %d lines want %d lines", nlgot, nlwant)
	}
	write(&sw, mode, klines)
	got := strings.TrimSpace(sw.String())
	want := strings.TrimSpace(`
runtime/coverage/apis.go:156.156,2.38 1 0
runtime/coverage/apis.go:183.185,18.3 1 1
runtime/coverage/emit.go:496.497,39.32 1 0
runtime/coverage/emit.go:565.565,4.15 1 99
`)
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s\n", got, want)
	}
}
