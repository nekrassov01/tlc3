package tlc3

import (
	"crypto/tls"
	"reflect"
	"testing"
)

func TestOutputType_String(t *testing.T) {
	tests := []struct {
		name string
		tr   OutputType
		want string
	}{
		{
			name: "json",
			tr:   OutputTypeJSON,
			want: "json",
		},
		{
			name: "prettyjson",
			tr:   OutputTypePrettyJSON,
			want: "prettyjson",
		},
		{
			name: "text",
			tr:   OutputTypeText,
			want: "text",
		},
		{
			name: "compressedtext",
			tr:   OutputTypeCompressedText,
			want: "compressedtext",
		},
		{
			name: "markdown",
			tr:   OutputTypeMarkdown,
			want: "markdown",
		},
		{
			name: "backlog",
			tr:   OutputTypeBacklog,
			want: "backlog",
		},
		{
			name: "tsv",
			tr:   OutputTypeTSV,
			want: "tsv",
		},
		{
			name: "none",
			tr:   OutputTypeNone,
			want: "none",
		},
		{
			name: "unknown",
			tr:   256,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("OutputType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		tr      OutputType
		want    []byte
		wantErr bool
	}{
		{
			name: "json",
			tr:   OutputTypeJSON,
			want: []byte(`"json"`),
		},
		{
			name: "prettyjson",
			tr:   OutputTypePrettyJSON,
			want: []byte(`"prettyjson"`),
		},
		{
			name: "text",
			tr:   OutputTypeText,
			want: []byte(`"text"`),
		},
		{
			name: "compressedtext",
			tr:   OutputTypeCompressedText,
			want: []byte(`"compressedtext"`),
		},
		{
			name: "markdown",
			tr:   OutputTypeMarkdown,
			want: []byte(`"markdown"`),
		},
		{
			name: "backlog",
			tr:   OutputTypeBacklog,
			want: []byte(`"backlog"`),
		},
		{
			name: "tsv",
			tr:   OutputTypeTSV,
			want: []byte(`"tsv"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("OutputType.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OutputType.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOutputType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    OutputType
		wantErr bool
	}{
		{
			name: "json",
			args: args{
				s: "json",
			},
			want:    OutputTypeJSON,
			wantErr: false,
		},
		{
			name: "prettyjson",
			args: args{
				s: "prettyjson",
			},
			want:    OutputTypePrettyJSON,
			wantErr: false,
		},
		{
			name: "text",
			args: args{
				s: "text",
			},
			want:    OutputTypeText,
			wantErr: false,
		},
		{
			name: "compressedtext",
			args: args{
				s: "compressedtext",
			},
			want:    OutputTypeCompressedText,
			wantErr: false,
		},
		{
			name: "markdown",
			args: args{
				s: "markdown",
			},
			want:    OutputTypeMarkdown,
			wantErr: false,
		},
		{
			name: "backlog",
			args: args{
				s: "backlog",
			},
			want:    OutputTypeBacklog,
			wantErr: false,
		},
		{
			name: "tsv",
			args: args{
				s: "tsv",
			},
			want:    OutputTypeTSV,
			wantErr: false,
		},
		{
			name: "unsupported",
			args: args{
				s: "unsupported",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOutputType(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOutputType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseOutputType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLSVersion_String(t *testing.T) {
	tests := []struct {
		name string
		tr   TLSVersion
		want string
	}{
		{
			name: "1.0",
			tr:   TLSVersion10,
			want: "1.0",
		},
		{
			name: "1.1",
			tr:   TLSVersion11,
			want: "1.1",
		},
		{
			name: "1.2",
			tr:   TLSVersion12,
			want: "1.2",
		},
		{
			name: "1.3",
			tr:   TLSVersion13,
			want: "1.3",
		},
		{
			name: "none",
			tr:   TLSVersionNone,
			want: "none",
		},
		{
			name: "unknown",
			tr:   256,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("TLSVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLSVersion_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		tr      TLSVersion
		want    []byte
		wantErr bool
	}{
		{
			name: "1.0",
			tr:   TLSVersion10,
			want: []byte(`"1.0"`),
		},
		{
			name: "1.1",
			tr:   TLSVersion11,
			want: []byte(`"1.1"`),
		},
		{
			name: "1.2",
			tr:   TLSVersion12,
			want: []byte(`"1.2"`),
		},
		{
			name: "1.3",
			tr:   TLSVersion13,
			want: []byte(`"1.3"`),
		},
		{
			name: "none",
			tr:   TLSVersionNone,
			want: []byte(`"none"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("TLSVersion.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TLSVersion.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTLSVersion(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    uint16
		wantErr bool
	}{
		{
			name: "1.0",
			args: args{
				s: "1.0",
			},
			want:    tls.VersionTLS10,
			wantErr: false,
		},
		{
			name: "1.1",
			args: args{
				s: "1.1",
			},
			want:    tls.VersionTLS11,
			wantErr: false,
		},
		{
			name: "1.2",
			args: args{
				s: "1.2",
			},
			want:    tls.VersionTLS12,
			wantErr: false,
		},
		{
			name: "1.3",
			args: args{
				s: "1.3",
			},
			want:    tls.VersionTLS13,
			wantErr: false,
		},
		{
			name: "unsupported",
			args: args{
				s: "unsupported",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTLSVersion(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTLSVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTLSVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
