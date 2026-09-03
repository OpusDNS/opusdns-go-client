// This module exists only to keep spec/ out of the published module.
//
// A directory containing a go.mod is excluded from the enclosing module's zip,
// so the 1.1 MB specification is not downloaded by everyone who runs
// `go get github.com/opusdns/opusdns-go-client`. It was about 40 percent of
// that download. Nothing here is importable and nothing imports it; the drift
// tests read the files by path, which module boundaries do not affect.
//
// Deleting this file silently puts the specification back into every download.
module github.com/opusdns/opusdns-go-client/spec

go 1.21
