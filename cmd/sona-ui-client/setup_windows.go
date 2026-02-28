//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func SetupHotkey(binding string) error {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	shellObj, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return fmt.Errorf("failed to create WScript.Shell: %w", err)
	}
	shell, _ := shellObj.QueryInterface(ole.IID_IDispatch)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	shortcutPath := os.Getenv("USERPROFILE") + "\\Desktop\\sona.lnk"

	cs, err := oleutil.CallMethod(shell, "CreateShortcut", shortcutPath)
	if err != nil {
		return fmt.Errorf("failed to create shortcut: %w", err)
	}

	shortcut := cs.ToIDispatch()

	oleutil.PutProperty(shortcut, "TargetPath", exe)
	oleutil.PutProperty(shortcut, "WorkingDirectory", filepath.Dir(exe))
	oleutil.PutProperty(shortcut, "Hotkey", binding)

	if _, err := oleutil.CallMethod(shortcut, "Save"); err != nil {
		return fmt.Errorf("failed to save shortcut: %w", err)
	}

	return nil
}
