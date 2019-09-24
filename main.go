/*
nmod provides support for operations on nested modules.

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
of the current directory (if it exists) and all modules in directories
recursively below the current directory.

dirs:
	mmod dirs [modules]

dirs prints the directories belonging to the given modules. Modules may be
supplied as space separated arguments. At least one module must be supplied.
*/
package main // import "github.com/jadekler/nmod"

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
prints the root directories of the current directory (if it exists) and all
modules in directories recursively below the current directory.

dirs:
	mmod dirs [modules]

dirs prints the directories belonging to the given modules. Modules may be
supplied as space separated arguments. If no modules are supplied, dirs
prints the directories belonging to the module of the current directory (if it
exists) and the directories of all modules in directories recursively below the
current directory.
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
		dirs := args
		if len(dirs) == 0 {
			// Look up and down for modfiles.
			modFiles, err := modFilesRecursivelyDown()
			if err != nil {
				return err
			}
			upwardsModFile, err := searchUpwardsForModule(".")
			if err != nil {
				return err
			}
			if upwardsModFile != "" {
				modFiles = append(modFiles, upwardsModFile)
			}

			// Aggregate the dirs of the modfiles.
			for _, modFile := range modFiles {
				dirs = append(dirs, filepath.Dir(modFile))
			}
		}
		return modules(dirs)
	case "help":
		usage()
	default:
	}

	// For both "dirs" and "rootdirs", we need to calculate modules:

	mods := args
	if len(mods) == 0 {
		// Look up and down for modfiles.
		modFiles, err := modFilesRecursivelyDown()
		if err != nil {
			return err
		}
		upwardsModFile, err := searchUpwardsForModule(".")
		if err != nil {
			return err
		}
		if upwardsModFile != "" {
			modFiles = append(modFiles, upwardsModFile)
		}

		// Aggregate the modules of the modfiles.
		for _, modFile := range modFiles {
			module, err := readModule(modFile)
			if err != nil {
				return err
			}
			mods = append(mods, module)
		}
	}

	switch cmd {
	case "rootdirs":
		return rootdirs(mods)
	case "dirs":
		return dirs(mods)
	default:
		usage()
	}

	return nil
}

func modules(dirs []string) error {
	modFiles := map[string]struct{}{}

	// First, look upwards for the modfile of each dir.
	for _, d := range dirs {
		// Pessimistically assume user didn't provide an absolute path - convert
		// every path into an absolute path.
		absD, err := filepath.Abs(d)
		if err != nil {
			return err
		}
		d = absD

		// Go up from specified directory until we see a go.mod.
		m, err := searchUpwardsForModule(d)
		if err != nil {
			return err
		}
		if m == "" {
			return fmt.Errorf("%s doesn't have a go.mod, nor do any of the directories above it", d)
		}
		modFiles[m] = struct{}{}
	}

	// Next, read each modfile and record its module.
	modules := map[string]struct{}{}
	for f := range modFiles {
		m, err := readModule(f)
		if err != nil {
			return err
		}
		modules[m] = struct{}{}
	}

	// Finally, print each module.
	for m := range modules {
		fmt.Println(m)
	}

	return nil
}

func rootdirs(dirs []string) error {
	// Gather (and de-dupe) all the go.mod files.
	modFiles := map[string]struct{}{}
	for _, d := range dirs {
		modFile, err := searchUpwardsForModule(d)
		if err != nil {
			return err
		}
		modFiles[modFile] = struct{}{}
	}

	// Report the modfile directories.
	for modFile := range modFiles {
		fmt.Println(filepath.Dir(modFile))
	}

	return nil
}

func dirs(args []string) error {
	// TODO(deklerk): implement
	return nil
}

func modFilesRecursivelyDown() ([]string, error) {
	var modFiles []string
	dedupedModFiles := map[string]struct{}{}
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "go.mod" {
			dedupedModFiles[path] = struct{}{}
		}
		return nil
	}); err != nil {
		return modFiles, err
	}

	for modFile := range dedupedModFiles {
		modFiles = append(modFiles, modFile)
	}
	return modFiles, nil
}

// searchUpwardsForModule searches each directory above the given startDir for
// a go.mod file. It returns the file location of the go.mod. If no go.mod is
// found, it returns "", nil.
//
// TODO(deklerk): stops at the first go.mod, but really should keep going all
// the way.
func searchUpwardsForModule(startDir string) (string, error) {
	var absCurDir string
	for curDir := startDir; absCurDir != "/"; curDir += "/.." {
		var err error
		absCurDir, err = filepath.Abs(curDir)
		if err != nil {
			return "", err
		}
		modFile := absCurDir + "/go.mod"
		_, err = os.Stat(modFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		return modFile, nil
	}
	return "", nil
}

var moduleRegexp = regexp.MustCompile("module (.+)")

// readModule reads a given go.mod file and returns its module name.
func readModule(f string) (string, error) {
	outbytes, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	matches := moduleRegexp.FindAllStringSubmatch(string(outbytes), -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("%s doesn't seem to have a module declaration", f)
	}
	return matches[0][1], nil
}
