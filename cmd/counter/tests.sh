#!/usr/bin/env rc

make

# You can get stylus-interpreter from https://github.com/fluidity-money/9lives.so/tree/main/tools/stylus-interpreter

stylus-interpreter \
	--url https://testnet-rpc.superposition.so \
	go-on-stylus.wasm \
	`{cast calldata 'add(uint256)' 123}
