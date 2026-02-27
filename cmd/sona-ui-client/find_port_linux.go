//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func findPort(processName string) (string, error) {
	cmd := exec.Command("ss", "-ltnp")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	re := regexp.MustCompile(`(127\.0\.0\.1|\*):(\d+).*` + processName)

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches := re.FindStringSubmatch(line)
			return matches[2], nil
		}
	}

	return "", fmt.Errorf("port not found")
}
