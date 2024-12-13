tlc3
====

[![CI](https://github.com/nekrassov01/tlc3/actions/workflows/ci.yml/badge.svg)](https://github.com/nekrassov01/tlc3/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/tlc3)](https://goreportcard.com/report/github.com/nekrassov01/tlc3)
![GitHub](https://img.shields.io/github/license/nekrassov01/tlc3)
![GitHub](https://img.shields.io/github/v/release/nekrassov01/tlc3)

CLI application for checking TLS certificate information

Usage
-----

```text
NAME:
   tlc3 - TLS cert checker CLI

USAGE:
   tlc3 [global options] [arguments...]

VERSION:
   0.0.0

DESCRIPTION:
   CLI application for checking TLS certificate information

GLOBAL OPTIONS:
   --completion value, -c value                           completion scripts: bash|zsh|pwsh
   --loglevel value, -l value                             loglevels: debug|info|warn|error (default: "info") [$TLC3_LOGLEVEL]
   --domain value, -d value [ --domain value, -d value ]  domain:port separated by commas
   --file value, -f value                                 path to newline-delimited list of domains
   --output value, -o value                               output format: json|table|markdown|backlog (default: "json") [$TLC3_OUTPUT]
   --timeout value, -t value                              network timeout: ns|us|ms|s|m|h (default: 5s) [$TLC3_TIMEOUT]
   --insecure, -i                                         skip verification of the cert chain and host name (default: false)
   --no-timeinfo, -n                                      hide fields related to the current time in table output (default: false)
   --timezone value, -z value                             time zone for datetime fields (default: "Local") [$TLC3_TIMEZONE]
   --help, -h                                             show help
   --version, -v                                          print the version
```

Example
-------

```bash
# Pass domains separated by commas. Return in JSON by default
tlc3 -d example.com,www.example.com

# Pass by file path of newline-delimited list of domains.
tlc3 -l ./list.txt

# Return in non-escape text format table
tlc3 -d example.com,www.example.com -o table

# Return in markdown format table
tlc3 -d example.com,www.example.com -o markdown

# Return in backlog format table
tlc3 -d example.com,www.example.com -o backlog

# Hide fields related to the current time. Ignored for JSON format
tlc3 -d example.com,www.example.com -o markdown -n

# Override timeout value for TLS connection and IP lookup. Default is 5 seconds
tlc3 -d example.com,www.example.com -t 10s

# Change timezone from local to specified location
tlc3 -d example.com,www.example.com -z "Asia/Tokyo"
```

Benchmark
---------

[A quick benchmark](./benchmark_test.go) after improvement using connection pool

```text
$ make bench 
go test -run=^$ -bench=. -benchmem -count 5 -cpuprofile=cpu.prof -memprofile=mem.prof
goos: darwin
goarch: arm64
pkg: github.com/nekrassov01/tlc3
Benchmark-8       152959             13743 ns/op            1679 B/op         21 allocs/op
Benchmark-8        90865             14195 ns/op            1697 B/op         21 allocs/op
Benchmark-8        96865             13927 ns/op            1693 B/op         21 allocs/op
Benchmark-8       116883             13485 ns/op            1677 B/op         21 allocs/op
Benchmark-8       108783             12433 ns/op            1646 B/op         20 allocs/op
PASS
ok      github.com/nekrassov01/tlc3     11.248s
```

Warning
-------

`--insecure`,`-i` option can be used to skip verification of the certificate chain and host name. However, this risks exposure to man-in-the-middle attacks and should not be used unless it is clear that there is no problem.

If this option is used, y/n must be returned for the next question.

```bash
$ tlc3 -d example.com,www.example.com -i
? [WARNING] insecure flag skips verification of the certificate chain and hostname. skip it? [y/N]
```

If automation is required, this restriction can be removed by setting the environment variable.

```bash
export TLC3_NON_INTERACTIVE=true
```

Installation
------------

Install with homebrew

```sh
brew install nekrassov01/tap/tlc3
```

Install with go

```sh
go install github.com/nekrassov01/tlc3
```

Or download binary from [releases](https://github.com/nekrassov01/tlc3/releases)

Shell completion
----------------

Supported shells are as follows:

- bash
- zsh
- pwsh

```sh
tlc3 --completion bash|zsh|pwsh
```

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/tlc3/blob/main/LICENSE)
