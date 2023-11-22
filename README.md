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
   --domain value, -d value [ --domain value, -d value ]  domain:port separated by commas
   --list value, -l value                                 path to newline-delimited list of domains
   --output value, -o value                               output format: json|markdown|backlog (default: "json")
   --timeout value, -t value                              network timeout: ns|us|ms|s|m|h (default: "5s")
   --insecure, -i                                         skip verification of the cert chain and host name (default: false)
   --no-timeinfo, -n                                      hide fields related to the current time in table output (default: false)
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

# Return in markdown format table
tlc3 -d example.com,www.example.com -o markdown

# Return in backlog format table
tlc3 -d example.com,www.example.com -o backlog

# Hide fields related to the current time. Ignored for JSON format
tlc3 -d example.com,www.example.com -o markdown -n

# Override timeout value for TLS connection and IP lookup. Default is 5 seconds
tlc3 -d example.com,www.example.com -t 10s
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

Download binary from the release page or install it with the following command:

```sh
go install github.com/nekrassov01/tlc3
```

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
