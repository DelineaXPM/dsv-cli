package main

import (
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
)

// 🧪 [WIP] Brew runs the local test for the brew formula.
// Requires running on darwin system.
// TODO: Not yet functional
func (Test) Brew() error {
	pterm.Info.Printfln("Running brew test")
	if err := sh.RunWithV(map[string]string{
		"HOMEBREW_NO_AUTO_UPDATE": "1",
	},
		"brew", "install", "--build-from-source", "dsv-cli"); err != nil {
		pterm.Error.Printfln("brew test failed: %v", err)
		return err
	}
	pterm.Success.Printfln("brew test passed")
	return nil
}
