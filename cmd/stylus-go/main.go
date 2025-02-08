// stylus-go: Generate source code when called as a generator. Hooks the
// associated functions of the struct into the entrypoint. Looks for any files
// that include the comment "//stylus", and then operates on those files in
// the current directory.

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"unicode"

	"golang.org/x/crypto/sha3"
)

type (
	funcInfo struct {
		name string
		args []funcArg
	}

	funcArg struct {
		localtype, abitype string
	}
)

func main() {
	// First, we find a file that contains a struct that's marked as "stylus
	// entrypoint" as its comment. Then we start to look for associated
	// functions connected to this struct. For each function, we compute the
	// signature and use it to generate source code that generates the
	// entrypoint for the contract.
	fst := token.NewFileSet()
	pkgs, err := parser.ParseDir(fst, ".", nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		panic(err)
	}
	// We try to find a structure that serves as the entrypoint for this
	// code. We have to search the AST twice to find it before proceeding.
	var focusedStruct string
	// We use a buffer here so that we can source code check it before writing it.
	var explainedBuf bytes.Buffer
STRUCTSEARCH:
	for _, pkg := range pkgs {
		// First search to find the struct that has a comment stating "stylus entrypoint".
		for filename, n := range pkg.Files {
			for _, d := range n.Decls {
				decl, ok := d.(*ast.GenDecl)
				// If we're not a type and we don't have comments, skip.
				if !ok || decl.Tok != token.TYPE || decl.Doc == nil {
					continue
				}
				var containsCmt bool
				for _, c := range decl.Doc.List {
					if c.Text == "//stylus entrypoint" {
						containsCmt = true
						break
					}
				}
				if !containsCmt {
					continue STRUCTSEARCH
				}
				if focusedStruct != "" {
					// We keep searching anyway after finding the struct key so we might
					// find an instance of someone making a mistake as a precaution
					// against undefined behaviour with a mistake here.
					log.Fatal("finding struct: already found struct with entrypoint")
				}
				for _, spec := range decl.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						log.Fatalf("%v: mislabelled entrypoint, not type", filename)
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						log.Fatalf("%v: mislabelled entrypoint, not struct", filename)
					}
					// Converted fields to a stringable representation for creating calling the
					// setup func.
					var preambleFields []PreambleField
					for _, f := range st.Fields.List {
						if len(f.Names) == 0 {
							log.Fatalf("parse struct args: anon not allowed")
						}
						for _, n := range f.Names {
							x, err := exprToNewSetupFn(f.Type)
							if err != nil {
								log.Fatalf("parse struct args: bad expr: %v", err)
							}
							preambleFields = append(preambleFields, PreambleField{
								Name:    n.Name,
								SetupFn: x,
							})
						}
					}
					// It's time for us to spit out some code! We need some work here to set
					// up the entrypoint preamble, including reading every field in the struct,
					// noting the position, then calling the setup position if the code is correct.
					if err := OutputPreamble(&explainedBuf, ts.Name.Name, preambleFields...); err != nil {
						log.Fatalf("parse struct args: output preamble: %v", err)
					}
					// We need to remember this struct later to prevent us from finding
					// functions unrelated to it.
					focusedStruct = ts.Name.Name
				}
			}
		}
		if focusedStruct == "" {
			log.Fatal("no struct to generate with found")
		}
		for _, n := range pkg.Files {
			for _, d := range n.Decls {
				fn, ok := d.(*ast.FuncDecl)
				if !ok {
					continue
				}
				// Now that we found a function, check if it takes a pointer receiver for
				// a structure, and that it's type is defined.
				if fn.Recv == nil || fn.Type == nil || fn.Type.Params == nil {
					continue
				}
				// If the function's not titled, we want to assume they're not exporting this.
				if unicode.IsLower(rune(fn.Name.Name[0])) {
					continue
				}
				// If the function's not connected to our focused struct, we skip it.

				// Optionally explain the arguments by searching for a comment of "stylus
				// (uint256,?)+", which we can use to hint the appropriate pattern to
				// decode with later in the codegen. If the comment does not exist, then
				// we make a best effort here to guess. Once we know the local type
				// and the external-facing type, we can start to stub out the unpacking
				// of the byte array and the conversion in the generated code by using
				// one of the generated functions this package has in its internal path.
				var docList []*ast.Comment
				if fn.Doc != nil {
					docList = fn.Doc.List
				}
				// If the function has params, then we need to start translating them.
				var explainedArgs, convFns []string
				if fn.Type.Params != nil {
					var localArgs []string
					explainedArgs, localArgs, err = explainArgs(docList, fn.Type.Params.List)
					if err != nil {
						log.Fatalf("explain args: %v: %v", fn, err)
					}
					// Create the conversion functions now from a word:
					convFns = argsToConvFunctions(explainedArgs, localArgs)
				}
				// Now it's time for us to generate entrypoint code. Let's start by computing
				// the entrypoint receiver here.
				sel := createSelector(fn.Name.Name, explainedArgs...)
				// Time to finally spit out some code! We use the template to do this part.
				err = OutputMatching(&explainedBuf, fn.Name.Name, sel, convFns)
				if err != nil {
					log.Fatalf("generate functions: generate file: %v", err)
				}
			}
		}
	}
	// Looks like we're done! It's time to write the finale, then try to parse the generated code.
	fmt.Fprint(&explainedBuf, `
	return 1
}
`)
	testBuf := explainedBuf
	if _, err := parser.ParseFile(fst, "", &testBuf, parser.AllErrors); err != nil {
		explainedBuf.WriteTo(os.Stderr)
		panic(fmt.Sprintf("bad generated code (REPORTME): %v", err))
	}
	// Looks like everything went okay. We can print the generated code here.
	genF, err := os.OpenFile("stylus_generated.go", os.O_CREATE|os.O_WRONLY, 0770)
	if err != nil {
		log.Fatal("open file stylus_generated.go: ", err)
	}
	defer genF.Close()
	if _, err := explainedBuf.WriteTo(genF); err != nil {
		log.Fatalf("open file stylus_generated.go: ", err)
	}
}

// createSelector by taking the name and the args, then converting it to keccak256.
func createSelector(name string, args ...string) []byte {
	// Name should never be empty.
	n := []rune(name)
	n[0] = unicode.ToLower(n[0])
	b := sha3.NewLegacyKeccak256()
	fmt.Fprintf(b, "%s(%s)", string(n), strings.Join(args, ","))
	return b.Sum(nil)[:4]
}

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
