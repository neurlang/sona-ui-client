//go:build windows
// +build windows

package main

import (
	"github.com/atotto/clipboard"
)

func (smoke *smoke) copyToClipboard(text string) {
	clipboard.WriteAll(text)
}
