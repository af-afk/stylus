#!/usr/bin/env rc

make

stylus-interpreter --url https://testnet-rpc.superposition.so go-on-stylus.wasm `{cast calldata 'hello(uint256)' 123}
