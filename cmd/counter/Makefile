
go-on-stylus.wasm: $(shell find . -name '*.go') Makefile
	@tinygo build \
		-o go-on-stylus2.wasm \
		-target wasm-unknown \
		-gc leaking \
		-panic trap
	@wasm-opt \
		--dce \
		--rse \
		--signature-pruning \
		--strip-debug \
		--strip-producers \
		-Oz go-on-stylus2.wasm \
		-o go-on-stylus.wasm
	@rm go-on-stylus2.wasm
