package main

import "github.com/af-afk/stylus"

//go:generate stylus-go m

//stylus entrypoint
type Storage struct {}

const ErrorAlways = 100

func (s Storage) Revert() error {
	return stylus.Err(ErrorAlways, "I always revert!")
}

func main() {}
