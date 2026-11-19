package main

import (
	"fmt"
	"os"

	"github.com/zot/frictionless/applink"
)

// runLinkapp is `frictionless linkapp add|remove|list [APP] [--dir DIR]`. The
// mcp script's linkapp command hands off to it.
// CRC: crc-AppLink.md | R233
func runLinkapp(args []string) int {
	baseDir, rest := splitDirFlag(args)
	if err := applink.Command(rest, baseDir, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
