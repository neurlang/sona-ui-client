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

	// Example line:
	// LISTEN 0 128 127.0.0.1:39421 ... users:(("sona",pid=1234,fd=3))
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

func init() {
	port, err := findPort("sona")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Found port:", port)
}
