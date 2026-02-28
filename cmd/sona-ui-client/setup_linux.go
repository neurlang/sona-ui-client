//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SetupHotkey(binding string) (err1 error) {
	name := "sona"
	exe, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		exe = "./sona-ui-client"
	}
	cmd := filepath.Dir(exe) + "/sona-ui-client --once"

	base := "org.gnome.settings-daemon.plugins.media-keys"
	key := "/org/gnome/settings-daemon/plugins/media-keys/custom-keybindings/custom0/"

	args := [][]string{
		{"set", base, "custom-keybindings", "['" + key + "']"},
		{"set", base + ".custom-keybinding:" + key, "name", name},
		{"set", base + ".custom-keybinding:" + key, "command", cmd},
		{"set", base + ".custom-keybinding:" + key, "binding", binding},
	}

	for _, arg := range args {
		if err := exec.Command("gsettings", arg...).Run(); err != nil {
			fmt.Println("Error setting up hotkey:", "gsettings", arg," :", err)
			err1 = err
		}
	}
	return
}
