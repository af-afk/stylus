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
	"unicode"
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
				if len(fn.Recv.List) == 0 {
					continue
				}
				receiverName, err := exprToLocalType(fn.Recv.List[0].Type)
				if receiverName != focusedStruct && receiverName != "*" + focusedStruct {
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
		log.Fatal("open file stylus_generated.go: ", err)
	}
}
