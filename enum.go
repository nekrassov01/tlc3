package tlc3

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
)

// OutputType represents the output type of the renderer.
type OutputType int

const (
	OutputTypeNone           OutputType = iota // The output type that means none.
	OutputTypeJSON                             // The output type that means JSON format.
	OutputTypePrettyJSON                       // The output type that means pretty JSON format.
	OutputTypeText                             // The output type that means text format.
	OutputTypeCompressedText                   // The output type that means compressed text format.
	OutputTypeMarkdown                         // The output type that means markdown format.
	OutputTypeBacklog                          // The output type that means backlog format.
	OutputTypeTSV                              // The output type that means TSV format.
)

// String returns the string representation of the output type.
func (t OutputType) String() string {
	switch t {
	case OutputTypeNone:
		return "none"
	case OutputTypeJSON:
		return "json"
	case OutputTypePrettyJSON:
		return "prettyjson"
	case OutputTypeText:
		return "text"
	case OutputTypeCompressedText:
		return "compressedtext"
	case OutputTypeMarkdown:
		return "markdown"
	case OutputTypeBacklog:
		return "backlog"
	case OutputTypeTSV:
		return "tsv"
	default:
		return ""
	}
}

// MarshalJSON returns the JSON representation of the output type.
func (t OutputType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// ParseOutputType parses the output type from the string representation.
func ParseOutputType(s string) (OutputType, error) {
	switch s {
	case OutputTypeJSON.String():
		return OutputTypeJSON, nil
	case OutputTypePrettyJSON.String():
		return OutputTypePrettyJSON, nil
	case OutputTypeText.String():
		return OutputTypeText, nil
	case OutputTypeCompressedText.String():
		return OutputTypeCompressedText, nil
	case OutputTypeMarkdown.String():
		return OutputTypeMarkdown, nil
	case OutputTypeBacklog.String():
		return OutputTypeBacklog, nil
	case OutputTypeTSV.String():
		return OutputTypeTSV, nil
	default:
		return OutputTypeNone, fmt.Errorf("unsupported output type: %q", s)
	}
}

// TLSVersion represents the TLS version.
type TLSVersion int

const (
	TLSVersionNone TLSVersion = iota // The minimum TLS version that means none.
	TLSVersion10                     // The minimum TLS version 1.0.
	TLSVersion11                     // The minimum TLS version 1.1.
	TLSVersion12                     // The minimum TLS version 1.2.
	TLSVersion13                     // The minimum TLS version 1.3.
)

// String returns the string representation of the tls version.
func (t TLSVersion) String() string {
	switch t {
	case TLSVersionNone:
		return "none"
	case TLSVersion10:
		return "1.0"
	case TLSVersion11:
		return "1.1"
	case TLSVersion12:
		return "1.2"
	case TLSVersion13:
		return "1.3"
	default:
		return ""
	}
}

// MarshalJSON returns the JSON representation of the tls version.
func (t TLSVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// ParseTLSVersion parses the TLS version from the string representation.
func ParseTLSVersion(s string) (uint16, error) {
	switch s {
	case TLSVersion10.String():
		return tls.VersionTLS10, nil
	case TLSVersion11.String():
		return tls.VersionTLS11, nil
	case TLSVersion12.String():
		return tls.VersionTLS12, nil
	case TLSVersion13.String():
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("unsupported tls version: %q", s)
	}
}
