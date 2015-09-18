package main

import (
	"fmt"
	"os"

	"github.com/mgutz/ansi"
)

var successize func(string) string = ansi.ColorFunc("green+b")
var errize func(string) string = ansi.ColorFunc("red+b")
var heyYou func(string) string = ansi.ColorFunc("yellow+b")

func success(s string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", successize("Success:"), s)
}

func failure(s string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", errize("ERROR:"), s)
}

func fatal(s string) {
	failure(s)
	os.Exit(1)
}
