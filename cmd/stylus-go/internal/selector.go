package internal

import (
	"unicode"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

// createSelector by taking the name and the args, then converting it to keccak256.
func createSelector(name string, args ...string) []byte {
	// Name should never be empty.
	n := []rune(name)
	n[0] = unicode.ToLower(n[0])
	b := sha3.NewLegacyKeccak256()
	fmt.Fprintf(b, "%s(%s)", string(n), strings.Join(args, ","))
	return b.Sum(nil)[:4]
}