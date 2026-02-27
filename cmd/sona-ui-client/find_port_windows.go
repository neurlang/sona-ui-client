//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func findPort(processName string) (string, error) {
	cmd := exec.Command("netstat", "-ano")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	re := regexp.MustCompile(`127\.0\.0\.1:(\d+)\s+LISTENING\s+(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); matches != nil {
			port, pid := matches[1], matches[2]
			if isProcessName(pid, processName) {
				return port, nil
			}
		}
	}

	return "", fmt.Errorf("port not found")
}

func isProcessName(pid string, name string) bool {
	cmd := exec.Command("tasklist", "/FI", "PID eq "+pid, "/FO", "CSV")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), name)
}
