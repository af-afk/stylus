/*
stylus-go generates Stylus-compatible WASM code from Go files. It may be used to also
generate the WASM itself by using Tinygo from the command line as a backend.

It generates code that calls functions associated with a struct that's been explicitly
marked using "//stylus". It operates on files in the current directory, or as specified.

Return arguments that are *big.Int will be assumed to be int256, with the onus on the
programmer to do the conversion before the return type.
*/
package main
