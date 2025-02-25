package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/af-afk/stylus/cmd/stylus-go/internal"
)

func main() {
	if len(os.Args) == 1 {
		usage()
	}
	switch os.Args[1] {
	case "g", "ge", "gen":
		if args := os.Args[2:]; len(args) == 0 || args[0] != "-" {
			genDirs()
		} else {
			genStdin()
		}
	case "b", "bu", "bui", "build":
		build()
	case "m", "ma", "mak", "make":
		make()
	default:
		usage()
	}
}

// genDirs by generating for a directory.
func genDirs() {
	// We find a file that contains a struct that's marked as "stylus
	// entrypoint" as its comment. Then we start to look for associated
	// functions connected to this struct. For each function, we compute the
	// signature and use it to generate source code that generates the
	// entrypoint for the contract.
	fst := token.NewFileSet()
	dirs := os.Args[2:]
	if len(os.Args[2:]) == 0 {
		dirs = append(dirs, ".")
	}
	var files []internal.File
	for _, a := range dirs {
		pkgs, err := parser.ParseDir(fst, a, nil, parser.AllErrors|parser.ParseComments)
		if err != nil {
			log.Fatalf("parse dir %v: %v", a, err)
		}
		for _, p := range pkgs {
			for s, f := range p.Files {
				files = append(files, internal.File{s, f.Decls})
			}
		}
	}
	gen(fst, files...)
}

// genStdin by creating for only stdin.
func genStdin() {
	// We parse stdin, expecting it contains everything we want.
	if len(os.Args[2:]) > 1 {
		// Currently, we can't parse stdin and files at the same time.
		usage()
	}
	fst := token.NewFileSet()
	p, err := parser.ParseFile(fst, "stdin", os.Stdin, parser.AllErrors|parser.ParseComments)
	if err != nil {
		log.Fatalf("parse stdin: %v", err)
	}
	gen(fst, internal.File{"stdin", p.Decls})
}

func gen(fst *token.FileSet, files ...internal.File) {
	var buf bytes.Buffer
	if err := internal.Generate(&buf, os.Stderr, fst, files...); err != nil {
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

func build() {
	// This is a simple frontend for tinygo build with our preferred
	// arguments, so we simply construct the arguments, then use it with
	// os/exec. We can optionally use wasm-opt with an extra optional
	// argument (-wasm-opt).
	dontUseWasmopt := false
	outN := "contract.wasm"
	ignoreNext := false
	for i, a := range os.Args[2:] {
		switch a {
		case "-no-wasm-opt":
			dontUseWasmopt = true
		case "-o", "-out":
			// If -build is used as the last positional param, or -wasm-opt is the
			// next argument, we error with usage.
			if i+1 == len(os.Args[2:]) || os.Args[2+i] == "-wasm-opt" {
				usage()
			}
			outN = os.Args[2+i+1]
			ignoreNext = true
		default:
			if !ignoreNext {
				usage()
			}
		}
	}
	var err error
	outN, err = filepath.Abs(outN)
	if err != nil {
		log.Fatalf("abs out file %#v: ", outN, err)
	}
	tmpF, err := os.CreateTemp(filepath.Dir(outN), "tmp-*.wasm")
	if err != nil {
		log.Fatal("create temp: ", err)
	}
	tmpF.Close() // Close so we may use this with subcommands.
	tmpName := tmpF.Name()
	tinyGoArgs := []string{
		"build",
		"-target", "wasm-unknown",
		"-gc", "leaking",
		"-panic", "trap",
		"-o", tmpName,
	}
	// We make some assumptions wasm-opt won't have issues working on the
	// wasm file by directly using the out file.
	wasmOptArgs := []string{
		"--dce",
		"--rse",
		"--signature-pruning",
		"--strip-debug",
		"--strip-producers",
		"-Oz", tmpName,
		"-o", outN,
	}
	tinygoCmd := exec.Command("tinygo", tinyGoArgs...)
	tinygoCmd.Stdout = os.Stdout
	tinygoCmd.Stderr = os.Stderr
	if err := tinygoCmd.Err; err != nil {
		delF(tmpName)
		log.Fatal("tinygo cmd: ", err)
	}
	if err := tinygoCmd.Run(); err != nil {
		delF(tmpName)
		log.Fatal("tinygo run: ", err)
	}
	// Check if wasm-opt is in the PATH. If it's not, then we log that's the
	// case, and we don't continue.
	wasmOptPath, err := exec.LookPath("wasm-opt")
	if (err != nil || wasmOptPath == "") && !dontUseWasmopt {
		slog.Info("wasm-opt not installed, this build will be made without it. Install it using https://github.com/WebAssembly/binaryen for much smaller binaries, or disable this error notice with -no-wasm-opt", "err", err)
		dontUseWasmopt = true
	}
	if dontUseWasmopt {
		// Since the user elected not to use wasm-opt, we must move the temporary
		// file, then shut down.
		if err := os.Rename(tmpName, outN); err != nil {
			log.Fatalf("rename %#v to %#v: %v", tmpName, outN, err)
		}
		return
	}
	wasmoptCmd := exec.Command("wasm-opt", wasmOptArgs...)
	wasmoptCmd.Stdout = os.Stdout
	wasmoptCmd.Stderr = os.Stderr
	if err := wasmoptCmd.Err; err != nil {
		log.Fatal("wasm-opt cmd: ", err)
	}
	if err := wasmoptCmd.Run(); err != nil {
		log.Fatal("wasm-opt run: ", err)
	}
	// Looks like we're done. We need to remove the old temporary file.
	delF(tmpName)
}

func delF(n string) {
	if err := os.Remove(n); err != nil {
		log.Fatal("remove tmpfile %#v: %v", n, err)
	}
}

func make() {
	genDirs()
	build()
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage of %s: [[m[ake]]|[g[en] [-|path...]]|[b[uild] [-no-wasm-opt] [-o contract.wasm]]]

Commands:
gen: Create a new generated file from the source code.
build: Compile an already generated file using Tinygo.
make: Create a new generated file, then compile it.

Compiles to contract.wasm by default.
`,
		os.Args[0],
	)
	os.Exit(1)
}
