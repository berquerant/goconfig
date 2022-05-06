package main

import (
	"os"
)

type Rule int

const (
	Market Rule = iota
	Society
	Universe
	None
)

func check(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}

func main() {
	c := NewBuilder().
		Size(10).
		Rule(None).
		Reverse(Universe).
		Reader(nil).
		Build()

	check(c.Size.Default() == 10, "default size")
	check(c.Rule.Default() == None, "default rule")
	check(c.Reverse.Default() == Universe, "default reverse")
	check(c.Reader.Default() == nil, "default reader")

	check(c.Size.Get() == 10, "get default size")
	check(c.Rule.Get() == None, "get default rule")
	check(c.Reverse.Get() == Universe, "get default reverse")
	check(c.Reader.Get() == nil, "get default reader")

	check(!c.Size.IsModified(), "size is not modified")
	check(!c.Rule.IsModified(), "rule is not modified")
	check(!c.Reverse.IsModified(), "reverse is not modified")
	check(!c.Reader.IsModified(), "reader is not modified")

	c.Apply(
		WithSize(2),
		WithRule(Society),
		WithReverse(Market),
		WithReader(os.Stdin),
	)

	check(c.Size.IsModified(), "size is modified")
	check(c.Rule.IsModified(), "rule is modified")
	check(c.Reverse.IsModified(), "reverse is modified")
	check(c.Reader.IsModified(), "reader is modified")

	check(c.Size.Default() == 10, "default size")
	check(c.Rule.Default() == None, "default rule")
	check(c.Reverse.Default() == Universe, "default reverse")
	check(c.Reader.Default() == nil, "default reader")

	check(c.Size.Get() == 2, "get size")
	check(c.Rule.Get() == Society, "get rule")
	check(c.Reverse.Get() == Market, "get reverse")
	check(c.Reader.Get() == os.Stdin, "get reader")
}
