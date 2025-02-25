package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"go/ast"
	"go/parser"
	"go/token"
	"unicode"
	"io"
)

type File struct {
	Name  string
	Decls []ast.Decl
}

func Generate(out, werr io.Writer, fst *token.FileSet, files ...File) error {
	if len(files) == 0 {
		return fmt.Errorf("no files passed")
	}
	// We try to find a structure that serves as the entrypoint for this
	// code. We have to search the AST twice to find it before proceeding.
	var focusedStruct string
	// We use a buffer here so that we can source code check it before writing it.
	var explainedBuf bytes.Buffer
STRUCTSEARCH:
	// First search to find the struct that has a comment stating "stylus entrypoint".
	for filename, n := range files {
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
				// We keep searching anyway after finding the struct key, so we might
				// find an instance of someone making a mistake as a precaution
				// against undefined behaviour with a mistake here.
				return fmt.Errorf("finding struct: already found struct with entrypoint")
			}
			for _, spec := range decl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					return fmt.Errorf("%v: mislabelled entrypoint, not type", filename)
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					return fmt.Errorf("%v: mislabelled entrypoint, not struct", filename)
				}
				// Converted fields to a stringable representation for creating calling the
				// setup func.
				var preambleFields []PreambleField
				for _, f := range st.Fields.List {
					if len(f.Names) == 0 {
						return fmt.Errorf("parse struct args: anon not allowed")
					}
					for _, n := range f.Names {
						x, err := exprToNewSetupFn(f.Type)
						if err != nil {
							return fmt.Errorf("parse struct args: bad expr: %v", err)
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
				err := OutputPreamble(&explainedBuf, ts.Name.Name, preambleFields...)
				if err != nil {
					return fmt.Errorf("parse struct args: output preamble: %v", err)
				}
				// We need to remember this struct later to prevent us from finding
				// functions unrelated to it.
				focusedStruct = ts.Name.Name
			}
		}
		if focusedStruct == "" {
			return fmt.Errorf("no struct to generate with found")
		}
		for fname, n := range files {
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
				if len(fn.Recv.List) == 0 {
					continue
				}
				receiverName, err := exprToLocalType(fn.Recv.List[0].Type)
				if receiverName != focusedStruct && receiverName != "*"+focusedStruct {
					continue
				}
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
						return fmt.Errorf("explain args: %v: %v: %v", fname, fn, err)
					}
					// Create the conversion functions now from a word:
					convFns = argsToConvFunctions(explainedArgs, localArgs)
				}
				// Let's also check the arguments that this function takes, and spit out
				// some return type conversion functions if they're needed. We always expect
				// the form (something, error), so any variations of this should cause an error
				// here.
				if fn.Type.Results.List == nil {
					return fmt.Errorf("body results: %v: nil", fname)
				}
				returnFns, hasErrReturn, err := explainReturnFns(fn.Type.Results.List)
				if err != nil {
					log.Print(fn.Type.Results.List)
					return fmt.Errorf("explain return fns: %v: %v", fname, err)
				}
				// Now it's time for us to generate entrypoint code. Let's start by computing
				// the entrypoint receiver here.
				sel := createSelector(fn.Name.Name, explainedArgs...)
				// Time to finally spit out some code! We use the template to do this part.
				err = OutputMatching(
					&explainedBuf,
					fn.Name.Name,
					sel,
					convFns,
					returnFns,
					hasErrReturn,
				)
				if err != nil {
					return fmt.Errorf("generate functions: out buf: %v", err)
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
		explainedBuf.WriteTo(werr)
		return fmt.Errorf("bad generated code: %v", err)
	}
	if _, err := explainedBuf.WriteTo(out); err != nil {
		return fmt.Errorf("write buf: %v", err)
	}
	return nil
}
