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
