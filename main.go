package main

import (
	"fmt"
	ffi "github.com/aarshkshah1992/filecoin-ffi"
)

func main() {
	fooMessage := ffi.Message("hello foo")
	fooDigest := ffi.Hash(fooMessage)
	fmt.Println("fooDigest: ", fooDigest)
}
