package main

import (
	"fmt"
	"github.com/mattheath/base62"
	"testing"
)

func Test_toSha1(t *testing.T) {
	encodeInt64 := base62.EncodeInt64(32)
	fmt.Println(encodeInt64)
	fmt.Println(toSha1("sdadsadasda dsadsad "))
}
