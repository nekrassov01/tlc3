package tlc3

import (
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
