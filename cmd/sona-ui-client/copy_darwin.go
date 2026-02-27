//go:build darwin
// +build darwin

package main

import (
	"github.com/atotto/clipboard"
)

func (smoke *smoke) copyToClipboard(text string, callback func()) {
	clipboard.WriteAll(text)
	callback()
}
