package tlc3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func Test_GetCerts(t *testing.T) {
	type args struct {
		ctx      context.Context
		addrs    []string
		timeout  time.Duration
		location *time.Location
		insecure bool
	}
	tests := []struct {
		name    string
		args    args
		want    []*CertInfo
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				ctx:      context.Background(),
				addrs:    []string{addr},
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: true,
			},
			want:    basicExpects,
			wantErr: false,
		},
		{
			name: "utc",
			args: args{
				ctx:      context.Background(),
				addrs:    []string{addr},
				timeout:  5 * time.Second,
				location: time.UTC,
				insecure: true,
			},
			want:    utcExpects,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCerts(tt.args.ctx, tt.args.addrs, tt.args.timeout, tt.args.insecure, tt.args.location, DefaultTLSVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCerts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCerts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newConnector(t *testing.T) {
	type args struct {
		addr     string
		timeout  time.Duration
		location *time.Location
		insecure bool
	}
	tests := []struct {
		name    string
		args    args
		want    *connector
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				addr:     addr,
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: false,
			},
			want: &connector{
				addr:     addr,
				host:     host,
				port:     port,
				timeout:  5 * time.Second,
				location: time.Local,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: false, // #nosec G402
				},
			},
			wantErr: false,
		},
		{
			name: "complete host:port",
			args: args{
				addr:     "localhost",
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: false,
			},
			want: &connector{
				addr:     "localhost:443",
				host:     "localhost",
				port:     "443",
				timeout:  5 * time.Second,
				location: time.Local,
				config: &tls.Config{
					ServerName:         "localhost",
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: false, // #nosec G402
				},
			},
			wantErr: false,
		},
		{
			name: "error host:port",
			args: args{
				addr:     "localhost[",
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: false,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newConnector(tt.args.addr, tt.args.timeout, tt.args.insecure, tt.args.location, DefaultTLSVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("newConnector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil {
				return
			}
			if !reflect.DeepEqual(got.addr, tt.want.addr) {
				t.Errorf("addr = %v, want %v", got.addr, tt.want.addr)
			}
			if !reflect.DeepEqual(got.host, tt.want.host) {
				t.Errorf("host = %v, want %v", got.host, tt.want.host)
			}
			if !reflect.DeepEqual(got.port, tt.want.port) {
				t.Errorf("port = %v, want %v", got.port, tt.want.port)
			}
			if !reflect.DeepEqual(got.timeout, tt.want.timeout) {
				t.Errorf("timeout = %v, want %v", got.timeout, tt.want.timeout)
			}
			if !reflect.DeepEqual(got.location, tt.want.location) {
				t.Errorf("location = %v, want %v", got.location, tt.want.location)
			}
			if !reflect.DeepEqual(got.config, tt.want.config) {
				t.Errorf("config = %v, want %v", got.config, tt.want.config)
			}
		})
	}
}

func Test_connector_lookupIP(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr    string
		host    string
		port    string
		ips     []netip.Addr
		timeout time.Duration
		config  *tls.Config
		conn    *tls.Conn
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []netip.Addr
	}{
		{
			name: "basic",
			fields: fields{
				addr:    addr,
				host:    host,
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				config:  nil,
				conn:    nil,
			},
			args: args{
				ctx: ctx,
			},
			want: []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")},
		},
		{
			name: "empty",
			fields: fields{
				addr:    addr,
				host:    "dummy",
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				config:  nil,
				conn:    nil,
			},
			args: args{
				ctx: ctx,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMap.Delete(tt.fields.host)
			c := &connector{
				addr:    tt.fields.addr,
				host:    tt.fields.host,
				port:    tt.fields.port,
				ips:     tt.fields.ips,
				timeout: tt.fields.timeout,
				config:  tt.fields.config,
				conn:    tt.fields.conn,
			}
			c.lookupIP(tt.args.ctx)
			if !reflect.DeepEqual(c.ips, tt.want) {
				t.Errorf("lookupIP() = %v, want %v", c.ips, tt.want)
			}
		})
	}
}

func Test_connector_connect(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr    string
		host    string
		port    string
		ips     []netip.Addr
		timeout time.Duration
		config  *tls.Config
		conn    *tls.Conn
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				addr:    addr,
				host:    host,
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: true, // #nosec G402
				},
				conn: nil,
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				addr:    ":443",
				host:    host,
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: true, // #nosec G402
				},
				conn: nil,
			},
			args: args{
				ctx: ctx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connMap.Delete(tt.fields.host)
			c := &connector{
				addr:    tt.fields.addr,
				host:    tt.fields.host,
				port:    tt.fields.port,
				ips:     tt.fields.ips,
				timeout: tt.fields.timeout,
				config:  tt.fields.config,
				conn:    tt.fields.conn,
			}
			if err := c.connect(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("connector.connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connector_getCert(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr     string
		host     string
		port     string
		ips      []netip.Addr
		timeout  time.Duration
		location *time.Location
		config   *tls.Config
		conn     *tls.Conn
	}
	tests := []struct {
		name    string
		fields  fields
		want    *CertInfo
		wantErr bool
		hook    func(c *connector) error
	}{
		{
			name: "basic",
			fields: fields{
				addr:     addr,
				host:     host,
				port:     port,
				ips:      []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")},
				timeout:  5 * time.Second,
				location: time.Local,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want:    basicExpects[0],
			wantErr: false,
		},
		{
			name: "utc",
			fields: fields{
				addr:     addr,
				host:     host,
				port:     port,
				ips:      []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")},
				timeout:  5 * time.Second,
				location: time.UTC,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want:    utcExpects[0],
			wantErr: false,
		},
		{
			name: "invalid conn",
			fields: fields{
				addr:     addr,
				host:     host,
				port:     port,
				ips:      []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")},
				timeout:  5 * time.Second,
				location: time.UTC,
				config: &tls.Config{
					ServerName:         host,
					MinVersion:         DefaultTLSVersion,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want:    nil,
			wantErr: true,
			hook: func(c *connector) error {
				c.conn = nil
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connector{
				addr:     tt.fields.addr,
				host:     tt.fields.host,
				port:     tt.fields.port,
				ips:      tt.fields.ips,
				timeout:  tt.fields.timeout,
				location: tt.fields.location,
				config:   tt.fields.config,
				conn:     tt.fields.conn,
			}
			if err := c.connect(ctx); err != nil {
				t.Fatal(err)
			}
			if tt.hook != nil {
				if err := tt.hook(c); err != nil {
					t.Fatal(err)
				}
			}
			got, err := c.getCert()
			if (err != nil) != tt.wantErr {
				t.Errorf("connector.getCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connector.getCert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSANs(t *testing.T) {
	uri, err := url.Parse("example.com/")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		cert *x509.Certificate
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				cert: &x509.Certificate{
					DNSNames:       []string{"example.com", "www.example.com"},
					EmailAddresses: []string{"aaa@example.com", "bbb@example.com"},
					IPAddresses:    []net.IP{net.ParseIP("192.168.10.10"), net.ParseIP("192.168.10.20")},
					URIs:           []*url.URL{uri},
				},
			},
			want: []string{"example.com", "www.example.com", "aaa@example.com", "bbb@example.com", "192.168.10.10", "192.168.10.20", "example.com/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSANs(tt.args.cert); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSANs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDaysLeft(t *testing.T) {
	type args struct {
		notAfter time.Time
		now      time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "full year difference",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 365,
		},
		{
			name: "one day before new year",
			args: args{
				notAfter: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "one day difference",
			args: args{
				notAfter: time.Date(2023, 1, 2, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "end of year",
			args: args{
				notAfter: time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 364,
		},
		{
			name: "same day and time",
			args: args{
				notAfter: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "one second difference",
			args: args{
				notAfter: time.Date(2023, 1, 1, 9, 0, 1, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "less than one day",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 9, 0, 1, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "last second",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 8, 59, 59, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "leap year difference",
			args: args{
				notAfter: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				now:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 365,
		},
		{
			name: "end of leap year",
			args: args{
				notAfter: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
				now:      time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDaysLeft(tt.args.notAfter, tt.args.now); got != tt.want {
				t.Errorf("getDaysLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}
