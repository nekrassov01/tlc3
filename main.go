package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	if err := newApp().cli.RunContext(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
