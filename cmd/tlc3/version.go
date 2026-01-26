package main

import "fmt"

// Version is the current version of tlc3.
const version = "0.0.24"

// Revision is the git revision of tlc3.
var revision = ""

// version returns the version and revision of tlc3.
func getVersion() string {
	if revision == "" {
		return version
	}
	return fmt.Sprintf("%s (revision: %s)", version, revision)
}
