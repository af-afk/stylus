package internal

import (
	"fmt"
	"go/ast"
)

// explainReturnFns, converting them to an array of functions to convert
// the local type to the byte array return type we need.
func explainReturnFns(fields []*ast.Field) (fns string, hasErr bool, err error) {
	if l := len(fields); l > 2 {
		return "", false, fmt.Errorf("bad return type, len: %v", l)
	}
	hasSnd := false
	if len(fields) == 2 {
		sndF, err := exprToLocalType(fields[1].Type)
		if err != nil {
			return "", false, fmt.Errorf("explain second arg: %v", err)
		}
		if sndF != "error" {
			return "", false, fmt.Errorf("explain second arg: not error")
		}
		hasSnd = true
	}
	fstF := fields[0]
	v, err := exprToLocalType(fstF.Type)
	if err != nil {
		return "", false, fmt.Errorf("explain expr: %v", err)
	}
	switch v {
	case "[]byte":
		return "stylus.BytesIdentity", hasSnd, nil
	case "*big.Int":
		return "stylus.BigToInt256Bytes", hasSnd, nil
	case "stylus.Uint256":
		return "stylus.Uint256ToBytes", hasSnd, nil
	case "error":
		return "stylus.ErrIdentity", hasSnd, nil
	case "stylus.Address":
		return "stylus.AddressToBytes", hasSnd, nil
	default:
		// TODO: support encoding []byte and more.
		return "", hasSnd, fmt.Errorf("return type: bad type: %v", v)
	}
}
