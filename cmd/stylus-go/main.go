// stylus-go: Generate source code when called as a generator. Hooks the
// associated functions of the struct into the entrypoint. Looks for any files
// that include the comment "//stylus", and then operates on those files in
// the current directory.

package main

import (
	"bytes"
	_ "embed"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/af-afk/stylus/cmd/stylus-go/internal"
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
	var files []internal.File
	for _, p := range pkgs {
		for s, f := range p.Files {
			files = append(files, internal.File{s, f.Decls})
		}
	}
	var buf bytes.Buffer
	if err := internal.Generate(&buf, fst, files...); err != nil {
		log.Fatal("generate: ", err)
	}
	f, err := os.OpenFile("stylus_generated.go", os.O_CREATE|os.O_WRONLY, 0770)
	if err != nil {
		log.Fatal("open file stylus_generated.go: ", err)
	}
	defer f.Close()
	if _, err := buf.WriteTo(f); err != nil {
		log.Fatal("write file stylus_generated.go: ", err)
	}
}
