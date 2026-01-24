package tlc3

import (
	"reflect"
	"testing"
)

func Test_GetDomainsFromFile(t *testing.T) {
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
			got, err := GetDomainsFromFile(tt.args.fp)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDomainsFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDomainsFromFile() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
