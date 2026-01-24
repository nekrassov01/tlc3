tlc3
====

[![CI](https://github.com/nekrassov01/tlc3/actions/workflows/ci.yml/badge.svg)](https://github.com/nekrassov01/tlc3/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nekrassov01/tlc3)](https://goreportcard.com/report/github.com/nekrassov01/tlc3)
![GitHub](https://img.shields.io/github/license/nekrassov01/tlc3)
![GitHub](https://img.shields.io/github/v/release/nekrassov01/tlc3)

CLI application for checking TLS certificate informations

Usage
-----

```text
NAME:
   tlc3 - TLS cert checker CLI

USAGE:
   tlc3 [global options]

VERSION:
   0.0.0 (revision: XXXXXXX)

DESCRIPTION:
   CLI application for checking TLS certificate informations

GLOBAL OPTIONS:
   --log-level string, -l string                                set log level (default: "INFO") [$TLC3_LOG_LEVEL]
   --address string, -a string [ --address string, -a string ]  domain:port separated by commas
   --file string, -f string                                     path to newline-delimited list of addresses
   --output string, -o string                                   set output type (default: "text") [$TLC3_OUTPUT_TYPE]
   --timeout duration, -t duration                              set network timeout duration (default: 5s) [$TLC3_TIMEOUT]
   --insecure, -i                                               skip verification of the cert chain and host name
   --static, -s                                                 hide fields related to the current time in table output
   --timezone string, -z string                                 time zone for datetime fields (default: "Local") [$TLC3_TIMEZONE]
   --help, -h                                                   show help
   --version, -v                                                print the version
```

Example
-------

```bash
# Pass domains separated by commas. Return in JSON by default
tlc3 -a example.com,www.example.com

# Pass by file path of newline-delimited list of domains.
tlc3 -f ./list.txt

# Return in non-escape text format table
tlc3 -a example.com,www.example.com -o table

# Return in markdown format table
tlc3 -a example.com,www.example.com -o markdown

# Return in backlog format table
tlc3 -a example.com,www.example.com -o backlog

# Hide fields related to the current time. Ignored for JSON format
tlc3 -a example.com,www.example.com -o markdown -n

# Override timeout value for TLS connection and IP lookup. Default is 5 seconds
tlc3 -a example.com,www.example.com -t 10s

# Change timezone from local to specified location
tlc3 -a example.com,www.example.com -z "Asia/Tokyo"
```

Benchmark
---------

[A quick benchmark](./benchmark_test.go) after improvement using connection pool

```text
$ make bench
go test -bench=. -benchmem -count 5 -benchtime=10000x -cpuprofile=cpu.prof -memprofile=mem.prof
goos: darwin
goarch: arm64
pkg: github.com/nekrassov01/tlc3
cpu: Apple M2
Benchmark_Single-8         10000               857.9 ns/op          1176 B/op         16 allocs/op
Benchmark_Single-8         10000               845.2 ns/op          1177 B/op         16 allocs/op
Benchmark_Single-8         10000               923.5 ns/op          1176 B/op         16 allocs/op
Benchmark_Single-8         10000               889.3 ns/op          1176 B/op         16 allocs/op
Benchmark_Single-8         10000               918.0 ns/op          1176 B/op         16 allocs/op
Benchmark_Multiple-8       10000              6199 ns/op            2794 B/op         38 allocs/op
Benchmark_Multiple-8       10000              5662 ns/op            2770 B/op         38 allocs/op
Benchmark_Multiple-8       10000              5776 ns/op            2770 B/op         38 allocs/op
Benchmark_Multiple-8       10000              5746 ns/op            2770 B/op         38 allocs/op
Benchmark_Multiple-8       10000              5745 ns/op            2770 B/op         38 allocs/op
PASS
ok      github.com/nekrassov01/tlc3     6.502s
```

Warning
-------

`--insecure`,`-i` option can be used to skip verification of the certificate chain and host name. However, this risks exposure to man-in-the-middle attacks and should not be used unless it is clear that there is no problem.

If this option is used, y/n must be returned for the next question.

```bash
$ tlc3 -a example.com,www.example.com -i
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

```sh
tlc3 completion bash|zsh|pwsh|fish
```

Author
------

[nekrassov01](https://github.com/nekrassov01)

License
-------

[MIT](https://github.com/nekrassov01/tlc3/blob/main/LICENSE)
