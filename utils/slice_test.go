package utils

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	stri := "hello world"
	bytes := String2Bytes(stri)
	fmt.Println(cap(bytes))
	fmt.Println(bytes[0:1])
}

type A struct {
	Age int
}

func TestStructInit(t *testing.T) {
	var a A
	a.Age = 1
}