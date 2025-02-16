# Go on Stylus

Go on Stylus is a code generator for compiling [Tinygo](https://tinygo.org) code to
Stylus-compatible WASM. It does so using a code generator.

## Example contract

The following contract (in an expanded form at cmd/counter/entry.go) simply adds, or
subtracts:

```go
//go:generate stylus-go gen

import (
	"math/big"
	"fmt"

	"github.com/af-afk/stylus"
)

//stylus entrypoint
type Storage struct {
	Counter stylus.StorageUint256
}

//stylus uint256
func (s Storage) Add(x *big.Int) (y *big.Int, error) {
	y = s.Counter.Get()
	y.Add(y, x)
	s.Counter.Set(y)
	return
}

func (s Storage) Sub(x *big.Int) (y *big.Int, err error) {
	y = s.Counter.Get()
	y.Sub(y, x)
	s.Counter.Set(y)
	return
}
```

Running `go generate` parses the files in the directory, and spits out a generated file
that can be compiled with Tinygo. Compilation with Tinygo can be done using `go-on-stylus`
as a frontend with `stylus-go build`.
