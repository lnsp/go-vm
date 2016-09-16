package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/lnsp/pkginfo"
)

var (
	pkg = pkginfo.PackageInfo{
		Name: "gvm",
		Version: pkginfo.PackageVersion{
			Major:      0,
			Minor:      1,
			Identifier: "dev",
		},
	}
)

func printVersion() {
	fmt.Println(pkg)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		printVersion()
		return
	}

	bytecode, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Errorf("%v\n", err)
		return
	}
	Boot(bytecode)
}
