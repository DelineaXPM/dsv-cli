package vaultcli

import (
	"strings"

	cst "github.com/DelineaXPM/dsv-cli/constants"
)

func ToFlagName(flag string) string {
	return strings.ReplaceAll(flag, ".", "-")
}

func GetFlagVal(flag string, args []string) string {
	shortFlag := cst.GetShortFlag(flag)

	flagNames := []string{"--" + ToFlagName(flag)} // Long form of the flag. For example: `--config`
	if shortFlag != "" {
		flagNames = append(flagNames, `-`+shortFlag) // Add short form of the flag. For example: `-c`
	}

	for i, arg := range args {
		for _, flagName := range flagNames {
			if arg == flagName && len(args)-1 >= i+1 {
				return args[i+1]
			}

			if strings.HasPrefix(arg, flagName+"=") {
				return arg[len(flagName)+1:]
			}
		}
	}

	return ""
}

// GetFilenameFromArgs tries to extract a filename from args. If args has a --data or -d flag and
// its value starts with an '@' followed by a filename, the function tries to capture that filename.
func GetFilenameFromArgs(args []string) string {
	var fileName string
	for i := range args {
		if args[i] == "--data" || args[i] == "-d" {
			if i+1 == len(args) {
				break
			}
			value := args[i+1]
			if strings.HasPrefix(value, cst.CmdFilePrefix) {
				fileName = value[1:]
			}
			break
		}
	}
	return fileName
}
