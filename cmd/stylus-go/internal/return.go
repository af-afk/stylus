package internal

import (
	"fmt"
	"go/ast"
)

// explainReturnFns, converting them to an array of functions to convert
// the local type to the byte array return type we need.
func explainReturnFns(fields []*ast.Field) (fns string, err error) {
	if len(fields) != 2 {
		return "", fmt.Errorf("bad return type")
	}
	sndF, err := exprToLocalType(fields[1].Type)
	if err != nil {
		return "", fmt.Errorf("explain second arg: %v", err)
	}
	if sndF != "error" {
		return "", fmt.Errorf("explain second arg: not error")
	}
	fstF := fields[0]
	v, err := exprToLocalType(fstF.Type)
	if err != nil {
		return "", fmt.Errorf("explain expr: %v", err)
	}
	switch v {
	case "[32]byte":
		return "stylus.BytesIdentity", nil
	case "*big.Int":
		return "stylus.BigToInt256Bytes", nil
	default:
		// TODO: support encoding []byte and more.
		return "", fmt.Errorf("return type: bad type: %v", v)
	}
}
