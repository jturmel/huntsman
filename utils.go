package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (m model) exportToCSV() (string, error) {
	if m.baseUrl == nil {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	downloadsDir := filepath.Join(home, "Downloads")
	if _, err := os.Stat(downloadsDir); os.IsNotExist(err) {
		downloadsDir = home
	}

	timestamp := time.Now().Format("20060102_150405")
	domain := strings.ReplaceAll(m.baseUrl.Host, ".", "-")
	filename := fmt.Sprintf("%s_%s.csv", timestamp, domain)
	filePath := filepath.Join(downloadsDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	_ = writer.Write([]string{"URL", "Status", "Type", "Size", "From Source"})

	for _, row := range m.table.Rows() {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	return filename, nil
}

func openURL(u string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", u}
	case "darwin":
		cmd = "open"
		args = []string{u}
	default:
		cmd = "xdg-open"
		args = []string{u}
	}
	_ = exec.Command(cmd, args...).Start()
}
