//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.Create("dont_escape.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintln(f, "package repository\n")
	fmt.Fprintln(f, "var dontEscape = [256]bool{")

	dontEscape := [256]bool{}
	for c := uint8('a'); c <= uint8('z'); c++ {
		dontEscape[c] = true
	}
	for c := uint8('A'); c <= uint8('Z'); c++ {
		dontEscape[c] = true
	}
	for c := uint8('0'); c <= uint8('9'); c++ {
		dontEscape[c] = true
	}

	dontEscape['$'] = true
	dontEscape['&'] = true
	dontEscape['+'] = true
	dontEscape['-'] = true
	dontEscape['.'] = true
	dontEscape[':'] = true
	dontEscape['='] = true
	dontEscape['@'] = true
	dontEscape['_'] = true
	dontEscape['~'] = true

	dontEscape['\\'] = true
	dontEscape['/'] = true

	for c, v := range dontEscape {
		if v {
			fmt.Fprintf(f, "\ttrue,  // '%c'\n", c)
		} else {
			fmt.Fprintf(f, "\tfalse, // 0x%X\n", c)
		}
	}
	fmt.Fprintln(f, "}")
}
