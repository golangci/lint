// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/mvdan/lint"

	"github.com/kisielk/gotool"

	"github.com/mvdan/interfacer"
	unparam "github.com/mvdan/unparam/check"
)

var tests = flag.Bool("tests", false, "include tests")

func main() {
	flag.Parse()
	if err := runLinters(flag.Args()...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var linters = [...]struct {
	name    string
	checker lint.Checker
}{
	{"unparam", &unparam.Checker{}},
	{"interfacer", &interfacer.Checker{}},
}

type metaChecker struct {
	wd string

	lprog *loader.Program
	prog  *ssa.Program
}

func runLinters(args ...string) error {
	paths := gotool.ImportPaths(args)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	c := &metaChecker{wd: wd}
	var conf loader.Config
	if _, err := conf.FromArgs(paths, *tests); err != nil {
		return err
	}
	if c.lprog, err = conf.Load(); err != nil {
		return err
	}
	prog := ssautil.CreateProgram(c.lprog, 0)
	prog.Build()
	for _, l := range linters {
		issues, err := l.checker.Check(c.lprog, prog)
		if err != nil {
			return err
		}
		c.printIssues(l.name, issues)
	}
	return nil
}

func (c *metaChecker) printIssues(name string, issues []lint.Issue) {
	for _, issue := range issues {
		fpos := c.lprog.Fset.Position(issue.Pos()).String()
		if strings.HasPrefix(fpos, c.wd) {
			fpos = fpos[len(c.wd)+1:]
		}
		fmt.Printf("%s: %s (%s)\n", fpos, issue.Message(), name)
	}
}
