//go:build darwin
// +build darwin

package main

import (
	"github.com/atotto/clipboard"
)

func (smoke *smoke) copyToClipboard(text string) {
	clipboard.WriteAll(text)
}
