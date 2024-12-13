package main

import (
	"bytes"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var input = []*certInfo{
	{
		DomainName:  host,
		AccessPort:  port,
		IPAddresses: []net.IP{},
		Issuer:      "CN=local test CA",
		CommonName:  "local test CA",
		SANs:        []string{},
		NotBefore:   getTime("2023-01-01T09:00:00+09:00", time.Local),
		NotAfter:    getTime("2025-01-01T09:00:00+09:00", time.Local),
		CurrentTime: getTime("2024-01-01T09:00:00+09:00", time.Local),
		DaysLeft:    365,
	},
}

func Test_fromList(t *testing.T) {
	type args struct {
		fp string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				fp: "testdata/1.txt",
			},
			want:    []string{"localhost:8443", "127.0.0.1:8443"},
			wantErr: false,
		},
		{
			name: "trim space",
			args: args{
				fp: "testdata/2.txt",
			},
			want:    []string{"localhost:8443", "127.0.0.1:8443"},
			wantErr: false,
		},
		{
			name: "skip line",
			args: args{
				fp: "testdata/3.txt",
			},
			want:    []string{"localhost:8443", "127.0.0.1:8443"},
			wantErr: false,
		},
		{
			name: "trim single quote",
			args: args{
				fp: "testdata/4.txt",
			},
			want:    []string{"localhost:8443", "127.0.0.1:8443"},
			wantErr: false,
		},
		{
			name: "trim double quote",
			args: args{
				fp: "testdata/5.txt",
			},
			want:    []string{"localhost:8443", "127.0.0.1:8443"},
			wantErr: false,
		},
		{
			name: "comma separated",
			args: args{
				fp: "testdata/6.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "0 byte file",
			args: args{
				fp: "testdata/7.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no file provided",
			args: args{
				fp: "",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromList(tt.args.fp)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func Test_out(t *testing.T) {
	type args struct {
		input  []*certInfo
		format string
		omit   bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				input:  input,
				format: formatJSON.String(),
				omit:   false,
			},
			want: `[
  {
    "DomainName": "localhost",
    "AccessPort": "8443",
    "IPAddresses": [],
    "Issuer": "CN=local test CA",
    "CommonName": "local test CA",
    "SANs": [],
    "NotBefore": "2023-01-01T09:00:00+09:00",
    "NotAfter": "2025-01-01T09:00:00+09:00",
    "CurrentTime": "2024-01-01T09:00:00+09:00",
    "DaysLeft": 365
  }
]
`,
			wantErr: false,
		},
		{
			name: "table",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   false,
			},
			want: `+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
`,
			wantErr: false,
		},
		{
			name: "markdown",
			args: args{
				input:  input,
				format: formatMarkdownTable.String(),
				omit:   false,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
|------------|------------|-------------|------------------|---------------|------|-------------------------------|-------------------------------|-------------------------------|----------|
| localhost  |       8443 | \-          | CN=local test CA | local test CA | \-   | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
`,
			wantErr: false,
		},
		{
			name: "backlog",
			args: args{
				input:  input,
				format: formatBacklogTable.String(),
				omit:   false,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |h
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
`,
			wantErr: false,
		},
		{
			name: "json+omit",
			args: args{
				input:  input,
				format: formatJSON.String(),
				omit:   true,
			},
			want: `[
  {
    "DomainName": "localhost",
    "AccessPort": "8443",
    "IPAddresses": [],
    "Issuer": "CN=local test CA",
    "CommonName": "local test CA",
    "SANs": [],
    "NotBefore": "2023-01-01T09:00:00+09:00",
    "NotAfter": "2025-01-01T09:00:00+09:00",
    "CurrentTime": "2024-01-01T09:00:00+09:00",
    "DaysLeft": 365
  }
]
`,
			wantErr: false,
		},
		{
			name: "table+omit",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   true,
			},
			want: `+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
`,
			wantErr: false,
		},
		{
			name: "markdown+omit",
			args: args{
				input:  input,
				format: formatMarkdownTable.String(),
				omit:   true,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |
|------------|------------|-------------|------------------|---------------|------|-------------------------------|-------------------------------|
| localhost  |       8443 | \-          | CN=local test CA | local test CA | \-   | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
		{
			name: "backlog+omit",
			args: args{
				input:  input,
				format: formatBacklogTable.String(),
				omit:   true,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |h
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
		{
			name: "error 1",
			args: args{
				input:  input,
				format: "",
				omit:   false,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "error 2",
			args: args{
				input:  input,
				format: "",
				omit:   true,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := out(tt.args.input, output, tt.args.format, tt.args.omit); (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if output.String() != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", output.String(), tt.want)
			}
		})
	}
}

func Test_toJSON(t *testing.T) {
	type args struct {
		input []*certInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				input: input,
			},
			want: `[
  {
    "DomainName": "localhost",
    "AccessPort": "8443",
    "IPAddresses": [],
    "Issuer": "CN=local test CA",
    "CommonName": "local test CA",
    "SANs": [],
    "NotBefore": "2023-01-01T09:00:00+09:00",
    "NotAfter": "2025-01-01T09:00:00+09:00",
    "CurrentTime": "2024-01-01T09:00:00+09:00",
    "DaysLeft": 365
  }
]
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := toJSON(tt.args.input, output); (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if output.String() != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", output.String(), tt.want)
			}
			if diff := cmp.Diff(output.String(), tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func Test_toTable(t *testing.T) {
	type args struct {
		input  []*certInfo
		format string
		omit   bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "table",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   false,
			},
			want: `+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+-------------------------------+----------+
`,
			wantErr: false,
		},
		{
			name: "markdown",
			args: args{
				input:  input,
				format: formatMarkdownTable.String(),
				omit:   false,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
|------------|------------|-------------|------------------|---------------|------|-------------------------------|-------------------------------|-------------------------------|----------|
| localhost  |       8443 | \-          | CN=local test CA | local test CA | \-   | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
`,
			wantErr: false,
		},
		{
			name: "backlog",
			args: args{
				input:  input,
				format: formatBacklogTable.String(),
				omit:   false,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |h
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST | 2024-01-01 09:00:00 +0900 JST |      365 |
`,
			wantErr: false,
		},
		{
			name: "table+omit",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   true,
			},
			want: `+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
+------------+------------+-------------+------------------+---------------+------+-------------------------------+-------------------------------+
`,
			wantErr: false,
		},
		{
			name: "markdown+omit",
			args: args{
				input:  input,
				format: formatMarkdownTable.String(),
				omit:   true,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |
|------------|------------|-------------|------------------|---------------|------|-------------------------------|-------------------------------|
| localhost  |       8443 | \-          | CN=local test CA | local test CA | \-   | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
		{
			name: "backlog+omit",
			args: args{
				input:  input,
				format: formatBacklogTable.String(),
				omit:   true,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                     | NotAfter                      |h
| localhost  |       8443 | -           | CN=local test CA | local test CA | -    | 2023-01-01 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			if err := toTable(tt.args.input, output, tt.args.format, tt.args.omit); (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if output.String() != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", output.String(), tt.want)
			}
		})
	}
}
