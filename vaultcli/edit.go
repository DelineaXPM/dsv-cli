package vaultcli

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cst "thy/constants"
	"thy/errors"
	"thy/format"
	"thy/utils"

	"github.com/spf13/viper"
	"golang.org/x/sys/execabs"
)

type SaveFunc func(data []byte) (resp []byte, err *errors.ApiError)

func EditData(data []byte, saveFunc SaveFunc, startErr *errors.ApiError, retry bool) (edited []byte, runErr *errors.ApiError) {
	viper.Set(cst.Output, string(format.File))
	dataFormatted, errString := format.FormatResponse(data, nil, viper.GetBool(cst.Beautify))
	viper.Set(cst.Output, string(format.StdOut))
	if errString != "" {
		return nil, errors.NewS(errString)
	}
	dataEdited, err := doEditData([]byte(dataFormatted), startErr)
	if err != nil {
		return nil, err
	}
	resp, postErr := saveFunc(dataEdited)
	if retry && postErr != nil {
		return EditData(dataEdited, saveFunc, postErr, true)
	}
	return resp, postErr
}

func doEditData(data []byte, startErr *errors.ApiError) (edited []byte, runErr *errors.ApiError) {
	editorCmd, getErr := getEditorCmd()
	if getErr != nil || editorCmd == "" {
		return nil, getErr
	}
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, cst.CmdRoot)
	if err != nil {
		return nil, errors.New(err).Grow("Error while creating temp file to edit data")
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Warning: failed to remove temporary file: '%s'\n%v", tmpFile.Name(), err)
		}
	}()

	if err := ioutil.WriteFile(tmpFile.Name(), data, 0600); err != nil {
		return nil, errors.New(err).Grow("Error while copying data to temp file")
	}

	// This is necessary for Windows. Opening a file in a parent process and then
	// trying to write it in a child process is not allowed. So we close the file
	// in the parent process first. This does not affect the behavior on Unix.
	if err := tmpFile.Close(); err != nil {
		log.Printf("Warning: failed to close temporary file: '%s'\n%v", tmpFile.Name(), err)
	}

	editorPath, err := execabs.LookPath(editorCmd)
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Error while looking up path to editor %q", editorCmd))
	}
	args := []string{tmpFile.Name()}
	if startErr != nil && (strings.HasSuffix(editorPath, "vim") || strings.HasSuffix(editorPath, "vi")) {
		args = append(args, "-c")
		errMsg := fmt.Sprintf("Error saving to %s. Please correct and save again or exit: %s", cst.ProductName, startErr.String())
		args = append(args, fmt.Sprintf(`:echoerr '%s'`, strings.Replace(errMsg, `'`, `''`, -1)))
	}
	cmd := execabs.Command(editorPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Command failed to start: '%s %s'", editorCmd, tmpFile.Name()))
	}
	err = cmd.Wait()
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Command failed to return: '%s %s'", editorCmd, tmpFile.Name()))
	}
	edited, err = ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, errors.New(err).Grow(fmt.Sprintf("Failed to read edited file: %q", tmpFile.Name()))
	}
	if utils.SlicesEqual(data, edited) {
		return nil, errors.NewS("Data not modified")
	}
	if len(edited) == 0 {
		return nil, errors.NewS("Cannot save empty file")
	}

	return edited, startErr
}

func getEditorCmd() (string, *errors.ApiError) {
	if utils.NewEnvProvider().GetOs() == "windows" {
		return "notepad.exe", nil
	}
	editor := viper.GetString(cst.Editor)
	// if editor specified in cli-config
	if editor != "" {
		return editor, nil
	}

	// try to find default editor on system
	out, err := execabs.Command("bash", "-c", getDefaultEditorSh).Output()
	editor = strings.TrimSpace(string(out))
	if err != nil || editor == "" {
		return "", errors.New(err).Grow("Failed to find default text editor. Please set 'editor' in the cli-config or make sure $EDITOR, $VISUAL is set on your system.")
	}

	// verbose - let them know why a certain editor is being implicitly chosen
	log.Printf("Using editor '%s' as it is found as default editor on the system. To override, set in cli-config (%s config update editor <EDITOR_NAME>)", editor, cst.CmdRoot)
	return editor, nil
}

const getDefaultEditorSh = `
#!/bin/sh
if [ -n "$VISUAL" ]; then
  echo $VISUAL
elif [ -n "$EDITOR" ]; then
  echo $EDITOR
elif type sensible-editor >/dev/null 2>/dev/null; then
  echo sensible-editor "$@"
elif cmd=$(xdg-mime query default ) 2>/dev/null; [ -n "$cmd" ]; then
  echo "$cmd"
else
  editors='nano joe vi'
  if [ -n "$DISPLAY" ]; then
    editors="gedit kate $editors"
  fi
  for x in $editors; do
    if type "$x" >/dev/null 2>/dev/null; then
	  echo "$x"
	  exit 0
    fi
  done
fi
`
