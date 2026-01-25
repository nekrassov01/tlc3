package tlc3

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	type args struct {
		w          io.Writer
		data       []*CertInfo
		outputType OutputType
		static     bool
	}
	tests := []struct {
		name  string
		args  args
		want  *Renderer
		wantW string
	}{
		{
			name: "basic",
			args: args{
				w:          &bytes.Buffer{},
				data:       renderInput,
				outputType: OutputTypeJSON,
				static:     false,
			},
			want: &Renderer{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeJSON,
				static:     false,
				w:          &bytes.Buffer{},
			},
		},
		{
			name: "empty",
			args: args{
				w:          &bytes.Buffer{},
				data:       nil,
				outputType: OutputTypeJSON,
				static:     false,
			},
			want: &Renderer{
				Header:     header,
				Data:       nil,
				OutputType: OutputTypeJSON,
				static:     false,
				w:          &bytes.Buffer{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := NewRenderer(w, tt.args.data, tt.args.outputType, tt.args.static); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRenderer() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("NewRenderer() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func TestRenderer_String(t *testing.T) {
	type fields struct {
		Data       []*CertInfo
		OutputType OutputType
		w          io.Writer
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "basic",
			fields: fields{
				Data:       renderInput,
				OutputType: OutputTypeJSON,
				w:          &bytes.Buffer{},
			},
			want: `{
  "Header": [
    "DomainName",
    "AccessPort",
    "IPAddresses",
    "Issuer",
    "CommonName",
    "SANs",
    "NotBefore",
    "NotAfter",
    "CurrentTime",
    "DaysLeft"
  ],
  "Data": [
    {
      "DomainName": "localhost",
      "AccessPort": "8443",
      "IPAddresses": [
        "127.0.0.1",
        "::1"
      ],
      "Issuer": "CN=local test CA",
      "CommonName": "local test CA",
      "SANs": [
        "localhost",
        "127.0.0.1"
      ],
      "NotBefore": "2024-01-01T09:00:00+09:00",
      "NotAfter": "2025-01-02T09:00:00+09:00",
      "CurrentTime": "2025-01-01T09:00:00+09:00",
      "DaysLeft": 1
    }
  ],
  "OutputType": "json"
}`,
		},
		{
			name: "empty",
			fields: fields{
				Data:       nil,
				OutputType: OutputTypeText,
				w:          &bytes.Buffer{},
			},
			want: `{
  "Header": [
    "DomainName",
    "AccessPort",
    "IPAddresses",
    "Issuer",
    "CommonName",
    "SANs",
    "NotBefore",
    "NotAfter",
    "CurrentTime",
    "DaysLeft"
  ],
  "Data": null,
  "OutputType": "text"
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ren := &Renderer{
				Header:     header,
				Data:       tt.fields.Data,
				OutputType: tt.fields.OutputType,
				w:          tt.fields.w,
			}
			if got := ren.String(); got != tt.want {
				t.Errorf("Renderer.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Render(t *testing.T) {
	type fields struct {
		Header     []string
		Data       []*CertInfo
		OutputType OutputType
		static     bool
		w          io.Writer
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "json",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeJSON,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want:    `[{"DomainName":"localhost","AccessPort":"8443","IPAddresses":["127.0.0.1","::1"],"Issuer":"CN=local test CA","CommonName":"local test CA","SANs":["localhost","127.0.0.1"],"NotBefore":"2024-01-01T09:00:00+09:00","NotAfter":"2025-01-02T09:00:00+09:00","CurrentTime":"2025-01-01T09:00:00+09:00","DaysLeft":1}]`,
			wantErr: false,
		},
		{
			name: "prettyjson",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypePrettyJSON,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `[
  {
    "DomainName": "localhost",
    "AccessPort": "8443",
    "IPAddresses": [
      "127.0.0.1",
      "::1"
    ],
    "Issuer": "CN=local test CA",
    "CommonName": "local test CA",
    "SANs": [
      "localhost",
      "127.0.0.1"
    ],
    "NotBefore": "2024-01-01T09:00:00+09:00",
    "NotAfter": "2025-01-02T09:00:00+09:00",
    "CurrentTime": "2025-01-01T09:00:00+09:00",
    "DaysLeft": 1
  }
]`,
			wantErr: false,
		},
		{
			name: "text",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeText,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs      | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
| localhost  |       8443 | 127.0.0.1   | CN=local test CA | local test CA | localhost | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |        1 |
|            |            | ::1         |                  |               | 127.0.0.1 |                               |                               |                               |          |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
`,
			wantErr: false,
		},
		{
			name: "compressedtext",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeCompressedText,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs      | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
| localhost  |       8443 | 127.0.0.1   | CN=local test CA | local test CA | localhost | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |        1 |
|            |            | ::1         |                  |               | 127.0.0.1 |                               |                               |                               |          |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+-------------------------------+----------+
`,
			wantErr: false,
		},
		{
			name: "markdown",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeMarkdown,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `| DomainName | AccessPort | IPAddresses      | Issuer           | CommonName    | SANs                   | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |
|------------|------------|------------------|------------------|---------------|------------------------|-------------------------------|-------------------------------|-------------------------------|----------|
| localhost  |       8443 | 127.0.0.1<br>::1 | CN=local test CA | local test CA | localhost<br>127.0.0.1 | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |        1 |
`,
			wantErr: false,
		},
		{
			name: "backlog",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeBacklog,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `| DomainName | AccessPort | IPAddresses      | Issuer           | CommonName    | SANs                   | NotBefore                     | NotAfter                      | CurrentTime                   | DaysLeft |h
| localhost  |       8443 | 127.0.0.1&br;::1 | CN=local test CA | local test CA | localhost&br;127.0.0.1 | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST | 2025-01-01 09:00:00 +0900 JST |        1 |
`,
			wantErr: false,
		},
		{
			name: "tsv",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeTSV,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want: `DomainName	AccessPort	IPAddresses	Issuer	CommonName	SANs	NotBefore	NotAfter	CurrentTime	DaysLeft
localhost	8443	127.0.0.1,::1	CN=local test CA	local test CA	localhost,127.0.0.1	2024-01-01 09:00:00 +0900 JST	2025-01-02 09:00:00 +0900 JST	2025-01-01 09:00:00 +0900 JST	1
`,
			wantErr: false,
		},
		{
			name: "json+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeJSON,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want:    `[{"DomainName":"localhost","AccessPort":"8443","IPAddresses":["127.0.0.1","::1"],"Issuer":"CN=local test CA","CommonName":"local test CA","SANs":["localhost","127.0.0.1"],"NotBefore":"2024-01-01T09:00:00+09:00","NotAfter":"2025-01-02T09:00:00+09:00","CurrentTime":"2025-01-01T09:00:00+09:00","DaysLeft":1}]`,
			wantErr: false,
		},
		{
			name: "prettyjson+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypePrettyJSON,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `[
  {
    "DomainName": "localhost",
    "AccessPort": "8443",
    "IPAddresses": [
      "127.0.0.1",
      "::1"
    ],
    "Issuer": "CN=local test CA",
    "CommonName": "local test CA",
    "SANs": [
      "localhost",
      "127.0.0.1"
    ],
    "NotBefore": "2024-01-01T09:00:00+09:00",
    "NotAfter": "2025-01-02T09:00:00+09:00",
    "CurrentTime": "2025-01-01T09:00:00+09:00",
    "DaysLeft": 1
  }
]`,
			wantErr: false,
		},
		{
			name: "text+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeText,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs      | NotBefore                     | NotAfter                      |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
| localhost  |       8443 | 127.0.0.1   | CN=local test CA | local test CA | localhost | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST |
|            |            | ::1         |                  |               | 127.0.0.1 |                               |                               |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
`,
			wantErr: false,
		},
		{
			name: "compressedtext+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeCompressedText,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
| DomainName | AccessPort | IPAddresses | Issuer           | CommonName    | SANs      | NotBefore                     | NotAfter                      |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
| localhost  |       8443 | 127.0.0.1   | CN=local test CA | local test CA | localhost | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST |
|            |            | ::1         |                  |               | 127.0.0.1 |                               |                               |
+------------+------------+-------------+------------------+---------------+-----------+-------------------------------+-------------------------------+
`,
			wantErr: false,
		},
		{
			name: "markdown+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeMarkdown,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `| DomainName | AccessPort | IPAddresses      | Issuer           | CommonName    | SANs                   | NotBefore                     | NotAfter                      |
|------------|------------|------------------|------------------|---------------|------------------------|-------------------------------|-------------------------------|
| localhost  |       8443 | 127.0.0.1<br>::1 | CN=local test CA | local test CA | localhost<br>127.0.0.1 | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
		{
			name: "backlog+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeBacklog,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `| DomainName | AccessPort | IPAddresses      | Issuer           | CommonName    | SANs                   | NotBefore                     | NotAfter                      |h
| localhost  |       8443 | 127.0.0.1&br;::1 | CN=local test CA | local test CA | localhost&br;127.0.0.1 | 2024-01-01 09:00:00 +0900 JST | 2025-01-02 09:00:00 +0900 JST |
`,
			wantErr: false,
		},
		{
			name: "tsv+static",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: OutputTypeTSV,
				static:     true,
				w:          &bytes.Buffer{},
			},
			want: `DomainName	AccessPort	IPAddresses	Issuer	CommonName	SANs	NotBefore	NotAfter
localhost	8443	127.0.0.1,::1	CN=local test CA	local test CA	localhost,127.0.0.1	2024-01-01 09:00:00 +0900 JST	2025-01-02 09:00:00 +0900 JST
`,
			wantErr: false,
		},
		{
			name: "empty",
			fields: fields{
				Header:     header,
				Data:       renderInput,
				OutputType: 256,
				static:     false,
				w:          &bytes.Buffer{},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			ren := &Renderer{
				Header:     tt.fields.Header,
				Data:       tt.fields.Data,
				OutputType: tt.fields.OutputType,
				static:     tt.fields.static,
				w:          w,
			}
			if err := ren.Render(); (err != nil) != tt.wantErr {
				t.Errorf("Renderer.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(w.String(), tt.want) {
				t.Errorf("Renderer.Render() = %v, want %v", w.String(), tt.want)
			}
		})
	}
}
