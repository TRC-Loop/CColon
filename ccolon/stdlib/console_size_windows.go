//go:build windows

package stdlib

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func getTerminalSize() (width, height int) {
	cmd := exec.Command("cmd", "/c", "mode", "con")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80, 24
	}
	lines := strings.Split(string(out), "\n")
	w, h := 80, 24
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Columns") || strings.Contains(line, "Spalten") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					w = n
				}
			}
		}
		if strings.Contains(line, "Lines") || strings.Contains(line, "Zeilen") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					h = n
				}
			}
		}
	}
	return w, h
}
