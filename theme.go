package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

type Theme struct {
	FocusedColor    string `json:"focused_color"`
	BlurredColor    string `json:"blurred_color"`
	SpinnerColor    string `json:"spinner_color"`
	CheckMarkColor  string `json:"check_mark_color"`
	TableSelectedFG string `json:"table_selected_fg"`
	TableSelectedBG string `json:"table_selected_bg"`
}

func DefaultTheme() Theme {
	return Theme{
		FocusedColor:    "#bd93f9",
		BlurredColor:    "240",
		SpinnerColor:    "#bd93f9",
		CheckMarkColor:  "#bd93f9",
		TableSelectedFG: "229",
		TableSelectedBG: "#bd93f9",
	}
}

func LoadTheme() Theme {
	theme := DefaultTheme()

	// Try to load from theme.json in current directory, next to binary, or ~/.config/huntsman/theme.json
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	paths := []string{
		"theme.json",
		filepath.Join(exeDir, "theme.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "huntsman", "theme.json"),
	}

	// Add macOS specific path if on Darwin
	if runtime.GOOS == "darwin" {
		paths = append(paths, filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "huntsman", "theme.json"))
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err == nil {
				json.Unmarshal(data, &theme)
				break
			}
		}
	}

	return theme
}
