//go:build darwin
// +build darwin

package main

import (
	"errors"
	"fmt"
)

var ErrHotkeyUnsupported = errors.New(
	"global system hotkey setup is not supported on macOS; " +
	"macOS requires a background app to register hotkeys at runtime",
)

func SetupHotkey(binding string) error {
	return fmt.Errorf("SetupHotkey(%q): %w", binding, ErrHotkeyUnsupported)
}
