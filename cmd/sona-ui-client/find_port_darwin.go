//go:build darwin
// +build darwin

package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func findPort(processName string) (string, error) {
	cmd := exec.Command("netstat", "-anv")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	re := regexp.MustCompile(`(127\.0\.0\.1|\*)\.(\d)+\s+(\d+|\*\.\*)\s+LISTEN\s+\d+\s+\d+\s+(\S+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); matches != nil {
			port, procInfo := matches[2], strings.TrimSpace(matches[4])
			fields := strings.Fields(procInfo)
			for _, f := range fields {
				cmd := exec.Command("ps", "-p", f, "-o", "comm=")
				output, err := cmd.Output()
				if err == nil && strings.Contains(string(output), processName) {
					return port, nil
				}
			}
		}
	}

	return "", fmt.Errorf("port not found")
}
