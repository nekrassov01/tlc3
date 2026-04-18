package tlc3

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/nekrassov01/mintab"
)

var header = []string{
	"DomainName",
	"AccessPort",
	"IPAddresses",
	"Issuer",
	"CommonName",
	"SANs",
	"NotBefore",
	"NotAfter",
	"CurrentTime",
	"DaysLeft",
}

// Renderer represents the renderer for the output.
type Renderer struct {
	Header     []string
	Data       []*CertInfo
	OutputType OutputType

	static bool
	w      io.Writer
}

// NewRenderer creates a new renderer with the specified parameters.
func NewRenderer(w io.Writer, data []*CertInfo, outputType OutputType, static bool) *Renderer {
	return &Renderer{
		Header:     header,
		Data:       data,
		OutputType: outputType,

		static: static,
		w:      w,
	}
}

// String returns the string representation of the renderer.
func (ren *Renderer) String() string {
	b, _ := json.MarshalIndent(ren, "", "  ")
	return string(b)
}

// Render renders the output.
func (ren *Renderer) Render() error {
	switch ren.OutputType {
	case OutputTypeJSON, OutputTypePrettyJSON:
		return ren.toJSON()
	case OutputTypeText, OutputTypeCompressedText, OutputTypeMarkdown, OutputTypeBacklog:
		return ren.toTable()
	case OutputTypeTSV:
		return ren.toTSV()
	default:
		return nil
	}
}

func (ren *Renderer) toJSON() error {
	var b []byte
	var err error
	switch ren.OutputType {
	case OutputTypeJSON:
		b, err = json.Marshal(ren.Data)
	case OutputTypePrettyJSON:
		b, err = json.MarshalIndent(ren.Data, "", "  ")
	}
	if err != nil {
		return err
	}
	_, err = ren.w.Write(b)
	return err
}

func (ren *Renderer) toTable() error {
	opts := make([]mintab.Option, 0, 2)
	switch ren.OutputType {
	case OutputTypeText:
		opts = append(opts, mintab.WithFormat(mintab.TextFormat))
	case OutputTypeCompressedText:
		opts = append(opts, mintab.WithFormat(mintab.CompressedTextFormat))
	case OutputTypeMarkdown:
		opts = append(opts, mintab.WithFormat(mintab.MarkdownFormat))
	case OutputTypeBacklog:
		opts = append(opts, mintab.WithFormat(mintab.BacklogFormat))
	}
	if ren.static {
		opts = append(opts, mintab.WithIgnoreFields([]int{8, 9}))
	}
	table := mintab.New(ren.w, opts...)
	if err := table.Load(ren.toInput()); err != nil {
		return err
	}
	table.Render()
	return nil
}

func (ren *Renderer) toInput() mintab.Input {
	data := make([][]any, len(ren.Data))
	for i, row := range ren.Data {
		data[i] = []any{
			row.DomainName,
			row.AccessPort,
			row.IPAddresses,
			row.Issuer,
			row.CommonName,
			row.SANs,
			row.NotBefore,
			row.NotAfter,
			row.CurrentTime,
			row.DaysLeft,
		}
	}
	return mintab.Input{
		Header: header,
		Data:   data,
	}
}

func (ren *Renderer) toTSV() error {
	w := csv.NewWriter(ren.w)
	w.Comma = '\t'
	var h []string
	if ren.static {
		h = ren.Header[:8]
	} else {
		h = ren.Header
	}
	if err := w.Write(h); err != nil {
		return err
	}
	fields := make([]string, 0, 10)
	var ib, sb strings.Builder
	for _, row := range ren.Data {
		ib.Reset()
		for i, ip := range row.IPAddresses {
			if i > 0 {
				ib.WriteByte(',')
			}
			ib.WriteString(ip.String())
		}
		sb.Reset()
		for i, s := range row.SANs {
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(s)
		}
		fields = fields[:0]
		fields = append(fields,
			row.DomainName,
			row.AccessPort,
			ib.String(),
			row.Issuer,
			row.CommonName,
			sb.String(),
			row.NotBefore.String(),
			row.NotAfter.String(),
		)
		if !ren.static {
			fields = append(fields, row.CurrentTime.String(), strconv.Itoa(row.DaysLeft))
		}
		if err := w.Write(fields); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
