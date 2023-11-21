package main

import (
	"context"
	"fmt"
	"os"
)

const (
	Name     = "tlc3"
	Version  = "0.0.0"
	Revision = "HEAD"
)

func main() {
	if err := newApp().run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
