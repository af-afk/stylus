/*
stylus-go generates Stylus-compatible WASM code from Go files. It may be used to also
generate the WASM itself by using Tinygo from the command line as a backend.

It generates code that calls functions associated with a struct that's been explicitly
marked using "//stylus". It operates on files in the current directory, or as specified.
*/
package main
