package main

import (
	_ "embed"
	"fmt"
	"go/ast"
)

// exprToNewSetupFn by taking the name of the field in the struct, checking if
// it's one of ours, and if it is, then spitting out a New function
// equivalent of it, which should bump the runtime counter with where
// this storage offset lives.
func exprToNewSetupFn(t ast.Expr) (string, error) {
	n, err := exprToLocalType(t)
	if err != nil {
		return "", fmt.Errorf("stringify local type: %v", err)
	}
	switch n {
	case "stylus.StorageUint256":
		return "stylus.NewStorageUint256", nil
	default:
		return "", fmt.Errorf("new storage create: %v", err)
	}
}

// exprToLocalType by converting the local node to a human friendly type
// for code generation on a basis for each later.
func exprToLocalType(t ast.Expr) (string, error) {
	switch v := t.(type) {
	case *ast.StarExpr:
		x, err := exprToLocalType(v.X)
		if err != nil {
			return "", err
		}
		return "*" + x, nil
	case *ast.Ident:
		return v.Name, nil
	case *ast.ArrayType:
		x, err := exprToLocalType(v.Elt)
		if err != nil {
			return "", err
		}
		return "[]" + x, nil
	case *ast.SelectorExpr:
		// Presumably, a special function will be found this instance that
		// converts a byte array to this type.
		x, err := exprToLocalType(v.X)
		if err != nil {
			return "", err
		}
		return x + "." + v.Sel.Name, nil
	default:
		return "", fmt.Errorf("invalid type: %T", t)
	}
}
