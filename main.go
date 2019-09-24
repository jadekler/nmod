/*
nmod provides support for operations on submodules.

Usage:
        nmod <command> [arguments]

The commands are:

modules         print the modules in this repository
dir             print the root directory of a module in this repository
dirs            print the directories of a module in this repository
*/

package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: nmod <command> [arguments]

The commands are:
	modules         print the modules in this repository
	dir             print the root directory of a module in this repository
	dirs            print the directories of a module in this repository
`)
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	if err := nmod(args[0], args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nmod: %s\n", err)
		os.Exit(1)
	}
}

func nmod(cmd string, args []string) error {
	return nil
}
