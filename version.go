package tlc3

import "fmt"

// version is the current version of application.
const version = "0.1.2"

// revision is the git revision of application.
var revision = ""

// Version returns the version and revision of application.
func Version() string {
	if revision == "" {
		return version
	}
	return fmt.Sprintf("%s (revision: %s)", version, revision)
}
