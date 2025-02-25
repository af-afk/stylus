package internal

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

func fmtByteArray(x []byte) string {
	var buf strings.Builder
	fmt.Fprint(&buf, "[]byte{")
	for _, b := range x {
		fmt.Fprintf(&buf, "%v,", b)
	}
	buf.WriteRune('}')
	return buf.String()
}

// OutputMatching template code to the generated file. Takes the local
// function to do conversion from, the selector computed as keccak256,
// all of the conversion functions needed for each argument, and whether
// the contract actually returns a separate field for error.
func OutputMatching(w io.Writer, localFn string, sel []byte, convFns []string, returnFn string, hasErrReturn bool) error {
	// TODO: handle different sized words in the calldata.
	cdlen := len(convFns) * 32
	s := fmtByteArray(sel)
	return TmplMatching.Execute(w, struct {
		LocalFn      string
		Sel          string
		CdLen        int
		ConvFns      []string
		ReturnFn     string
		HasErrReturn bool
	}{localFn, s, cdlen, convFns, returnFn, hasErrReturn})
}

var TmplMatching = template.Must(template.New("matching").Parse(`
	// {{.LocalFn }}
	if bytes.Equal(args[:4], {{.Sel}}) {
		var err error
		if len(args) != 4 + {{.CdLen}} {
			return 1
		}
		{{range $i, $e := .ConvFns}}x{{$i}}, err := stylus.{{$e}}(args[4+({{$i}}*32):4+32+({{$i}}*32)])
		if err != nil {
			return 1
		}
		{{end}}{{if .HasErrReturn}}rd{{.LocalFn}}, err :{{else}}err {{end}}= sr.{{.LocalFn}}({{range $i, $e := .ConvFns}}x{{$i}},{{end}})
		if err != nil {
			if d, ok := stylus.IsStylusErr(err); ok {
				stylus.Rd = d
			}
			return 1
		} else {
			return 0
		}
		if rc != 0 {
			return rc
		}
		{{if .HasErrReturn}}
		stylus.Rd = append(stylus.Rd, {{.ReturnFn}}(rd{{.LocalFn}})...){{end}}
	}`,
))
