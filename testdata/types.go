package main

import (
	"fmt"
	"os"
)

type Rule int

const (
	Market Rule = iota
	Society
	Universe
	None
)

func main() {
	c := NewBuilder().
		Size(10).
		Rule(None).
		Reverse(Universe).
		Reader(nil).
		Build()
	c.Apply(
		WithSize(2),
		WithRule(Society),
		WithReverse(Market),
		WithReader(os.Stdin),
	)
	fmt.Printf("%#v\n", c.Size.Get())
	fmt.Printf("%#v\n", c.Rule.Get())
	fmt.Printf("%#v\n", c.Reverse.Get())
	fmt.Printf("%#v\n", c.Reader.Get())
}
