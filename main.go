// +build ignore

// Command diff implements a basic, line-based diff tool.
//
// It's only intended for illustration purposes and thus strictly worse than
// the standard diff utility.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Merovius/diff"
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatal("usage: diff <old> <new>")
	}
	a, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}
	colorize(diff.LineDiff(a, b))
}

func colorize(d []diff.TextDelta) {
	prefix := map[diff.Op]string{
		diff.OpA:  "\033[31m-",
		diff.OpEq: "\033[0m ",
		diff.OpB:  "\033[32m+",
	}

	for _, δ := range d {
		fmt.Printf("%s %s\n", prefix[δ.Op], δ.Text)
	}
}
