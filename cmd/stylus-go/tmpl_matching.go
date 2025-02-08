package main

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

func OutputMatching(w io.Writer, localFn string, sel []byte, convFns []string) error {
	// TODO: handle different sized words in the calldata.
	cdlen := len(convFns) * 32
	s := fmtByteArray(sel)
	return TmplMatching.Execute(w, struct {
		LocalFn string
		Sel     string
		CdLen   int
		ConvFns []string
	}{
		LocalFn: localFn,
		Sel:     s,
		CdLen:   cdlen,
		ConvFns: convFns,
	})
}

var TmplMatching = template.Must(template.New("matching").Parse(`
	if bytes.Equal(args[:4], {{.Sel}}) {
		if len(args) != 4 + {{.CdLen}} {
			return 1
		}
		{{range $i, $e := .ConvFns}}x{{$i}}, err := stylus.{{$e}}(args[4+({{$i}}*32):4+32+({{$i}}*32)])
		if err != nil {
			return 1
		}
		{{end}}rd, err = sr.{{.LocalFn}}({{range $i, $e := .ConvFns}}x{{$i}},{{end}})
		if err != nil {
			return 1
		} else {
			return 0
		}
	}`))
