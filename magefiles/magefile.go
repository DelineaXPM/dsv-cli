// 🧙 Mage replaces makefiles, and is written in Go.
//
// For more detailed information on a task, you can run: mage -h <task> (such as mage -h azure:aksauth).
package main

import (
	//mage:import
	_ "github.com/sheldonhull/magetools/gittools"
	//mage:import
	_ "github.com/sheldonhull/magetools/gotools"
	//mage:import
	_ "github.com/sheldonhull/magetools/docgen"
	//mage:import
	_ "github.com/sheldonhull/magetools/precommit"
)
