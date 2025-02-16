package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"strings"
)

func argsToConvFunctions(explainedArgs, localArgs []string) (convFns []string) {
	convFns = make([]string, len(explainedArgs))
	for i, eArg := range explainedArgs {
		lArg := localArgs[i]
		switch {
		case "*big.Int" == lArg && "uint256" == eArg:
			convFns[i] = "BUint256ToBig"
		case "*big.Int" == lArg && "int256" == eArg:
			convFns[i] = "BInt256ToBig"
		default:
			panic(fmt.Sprintf("impl fix needed: %v: %v", eArg, lArg))
		}
	}
	return
}
func explainArgs(docs []*ast.Comment, args []*ast.Field) (explainedArgs []string, localArgs []string, err error) {
	// Let's try to match the local type with the estimated type from this
	// lookup. If there's anything we can't convert, then we'll report it
	// now. Let's try to expand the form where we might have multiple
	// arguments concenated together in their type.
	for _, a := range args {
		lt, err := exprToLocalType(a.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("local type lookup: %v", err)
		}
		for range a.Names {
			localArgs = append(localArgs, lt)
		}
	}
	for _, d := range docs {
		if strings.HasPrefix(d.Text, "//stylus") {
			// Match everything after "//stylus ".
			explainedArgs = strings.Split(d.Text, " ")[1:]
			if l := len(explainedArgs); l != len(localArgs) {
				return nil, nil, fmt.Errorf("explained args len: %v != %v", l, len(localArgs))
			}
			for i, s := range explainedArgs {
				// Try to match whether these are valid combinations.
				if err := isValidArgCombo(localArgs[i], s); err != nil {
					return nil, nil, fmt.Errorf("arguments explanation: %v != %v: %v", localArgs[i], s, err)
				}
			}
			return
		}
	}
	// If no-one's supplied the stylus prefix, then we can simply convert
	// the local type representation here to the external-facing form.
	explainedArgs = make([]string, len(localArgs))
	for i, s := range localArgs {
		if explainedArgs[i], err = argToPreferredType(s); err != nil {
			return nil, nil, fmt.Errorf("preferred type conv: %v", err)
		}
	}
	return
}

func isValidArgCombo(localArg, exportArg string) error {
	allowed := map[string][]string{
		"*big.Int":       {"uint256", "int256"},
		"stylus.Uint256": {"uint256"},
	}
	if _, ok := allowed[localArg]; !ok {
		return fmt.Errorf("allowed not found: %v", localArg)
	}
	for _, s := range allowed[localArg] {
		if s == exportArg {
			return nil
		}
	}
	return fmt.Errorf("exportarg not found")
}

func argToPreferredType(s string) (string, error) {
	switch s {
	case "*big.Int":
		return "int256", nil
	case "stylus.Uint256":
		return "uint256", nil
	default:
		return "", fmt.Errorf("bad type: %v", s)
	}
}
