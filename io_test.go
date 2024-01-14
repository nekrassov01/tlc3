package main

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = []*certInfo{
	{
		DomainName:  host,
		AccessPort:  port,
		IPAddresses: []string{},
		Issuer:      "CN=local test CA",
		CommonName:  "local test CA",
		SANs:        []string{},
		NotBefore:   "2023-01-01T09:00:00+09:00",
		NotAfter:    "2025-01-01T09:00:00+09:00",
		CurrentTime: "2024-01-01T09:00:00+09:00",
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
				t.Errorf("fromList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromList() = %v, want %v", got, tt.want)
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
]`,
			wantErr: false,
		},
		{
			name: "table",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   false,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |
|------------|------------|-------------|------------------|---------------|------|---------------------------|---------------------------|---------------------------|----------|
| localhost  |       8443 | N/A         | CN=local test CA | local test CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |
|------------|------------|-------------|----------------------------|-------------------------|------|---------------------------|---------------------------|---------------------------|----------|
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |h
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
]`,
			wantErr: false,
		},
		{
			name: "table+omit",
			args: args{
				input:  input,
				format: formatTextTable.String(),
				omit:   true,
			},
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                 | NotAfter                  |
|------------|------------|-------------|------------------|---------------|------|---------------------------|---------------------------|
| localhost  |       8443 | N/A         | CN=local test CA | local test CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  |
|------------|------------|-------------|----------------------------|-------------------------|------|---------------------------|---------------------------|
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  |h
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
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
			got, err := out(tt.args.input, tt.args.format, tt.args.omit)
			if (err != nil) != tt.wantErr {
				t.Errorf("out() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("out() = %v, want %v", got, tt.want)
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
]`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toJSON(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("toJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("toJSON() = %v, want %v", got, tt.want)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf(diff)
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |
|------------|------------|-------------|------------------|---------------|------|---------------------------|---------------------------|---------------------------|----------|
| localhost  |       8443 | N/A         | CN=local test CA | local test CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |
|------------|------------|-------------|----------------------------|-------------------------|------|---------------------------|---------------------------|---------------------------|----------|
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  | CurrentTime               | DaysLeft |h
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 | 2024-01-01T09:00:00+09:00 |      365 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs | NotBefore                 | NotAfter                  |
|------------|------------|-------------|------------------|---------------|------|---------------------------|---------------------------|
| localhost  |       8443 | N/A         | CN=local test CA | local test CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  |
|------------|------------|-------------|----------------------------|-------------------------|------|---------------------------|---------------------------|
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
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
			want: `| DomainName | AccessPort | IPAddresses | Issuer                     | CommonName              | SANs | NotBefore                 | NotAfter                  |h
| localhost  |       8443 | N/A         | CN=local&nbsp;test&nbsp;CA | local&nbsp;test&nbsp;CA | N/A  | 2023-01-01T09:00:00+09:00 | 2025-01-01T09:00:00+09:00 |
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toTable(tt.args.input, tt.args.format, tt.args.omit)
			if (err != nil) != tt.wantErr {
				t.Errorf("toTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("toTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
