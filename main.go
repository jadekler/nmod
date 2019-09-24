/*
nmod provides support for operations on submodules.

Usage: nmod <command> [args...]

NOTE: nmod is built to be run at the root of a repository. It does NOT query
for modules - it just scans directories in a straight line above, and
recursively below, the working directory.

The commands are:
	modules			print the modules of the given dirs
	rootdirs		print the root dirs of the given modules
	dirs			print the dirs of the given modules

modules:
	nmod modules [dirs...]

modules prints the modules of the dirs if they're supplied. Dirs may be supplied
as space separated arguments. If no dirs are supplied, modules prints the module
of the current directory (if it exists) and all modules in directories
recursively below the current directory.

rootdirs:
	mmod rootdirs [modules...]

rootdirs prints the root directories of the given modules. Modules may be
supplied as space separated arguments. If no modules are supplied, rootdirs
prints the root directory of the module of the current directory (if it exists).

dirs:
	mmod dirs [modules...]

dirs prints the directories belonging to the given modules. Modules may be
supplied as space separated arguments. If no modules are supplied, dirs prints
the root directory of the module of the current directory (if it exists).
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: nmod <command> [args...]

NOTE: nmod is built to be run at the root of a repository. It does NOT query
for modules - it just scans directories in a straight line above, and
recursively below, the working directory.

The commands are:
	modules			print the modules of the given dirs
	rootdirs		print the root dirs of the given modules
	dirs			print the dirs of the given modules

modules:
	nmod modules [dirs...]

modules prints the modules of the dirs if they're supplied. Dirs may be supplied
as space separated arguments. If no dirs are supplied, modules prints the module
of the current directory (if it exists) and all modules in directories
recursively below the current directory.

rootdirs:
	mmod rootdirs [modules...]

rootdirs prints the root directories of the given modules. Modules may be
supplied as space separated arguments. If no modules are supplied, rootdirs
prints the root directory of the module of the current directory (if it exists).

dirs:
	mmod dirs [modules...]

dirs prints the directories belonging to the given modules. Modules may be
supplied as space separated arguments. If no modules are supplied, dirs prints
the root directory of the module of the current directory (if it exists).
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
	switch cmd {
	case "modules":
		return modules(args)
	case "rootdirs":
		return rootdirs(args)
	case "dirs":
		return dirs(args)
	case "help":
		usage()
	default:
		usage()
	}

	return nil
}

var moduleRegexp = regexp.MustCompile("module (.+)")

func modules(args []string) error {
	var modFiles []string

	// Go up until we see a go.mod.
	var d string
	for cur := "."; d != "/"; cur += "/.." {
		var err error
		d, err = filepath.Abs(cur)
		if err != nil {
			return err
		}
		modFile := d + "/go.mod"
		_, err = os.Stat(modFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		modFiles = append(modFiles, modFile)
	}

	// Now go recursively down collecting go.mod files.
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "go.mod" {
			modFiles = append(modFiles, path)
		}
		return nil
	}); err != nil {
		return err
	}

	dedupedMatches := map[string]struct{}{}
	for _, f := range modFiles {
		outbytes, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		matches := moduleRegexp.FindAllStringSubmatch(string(outbytes), -1)
		if len(matches) == 0 {
			return fmt.Errorf("%s doesn't seem to have a module declaration", f)
		}
		dedupedMatches[matches[0][1]] = struct{}{}
	}

	for m := range dedupedMatches {
		fmt.Println(m)
	}

	return nil
}

func rootdirs(args []string) error {
	return nil
}

func dirs(args []string) error {
	return nil
}
