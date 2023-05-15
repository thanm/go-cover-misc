# go-cover-misc

Go language code coverage utilities.

## cmd/sortcovfuncs

A helper program for sorting the output of "go tool cover -func" (sort based on coverage percentage).

## cmd/pnormalize

A helper program for canonicalizing/normalizing legacy Go coverage profiles (of the format emitted by "go test -coverprofile=...").

## cmd/build-cover-tooldir/main.go

Small helper program for helping collect profiles from "-cover" built versions of Go tools (e.g. compile, link, etc). Figures out which tools to build, builds them (into single dir), then generates a toolexec wrapper that will pick them up.,
