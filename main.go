/*
Huntsman - A TUI app that spiders a website and lists all the resources it finds.

Copyright (C) 2026 Joshua Turmel

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#bd93f9"))
	blurredStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	maxConcurrency = 10
)

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL to spider..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 96
	ti.Prompt = " "

	fi := textinput.New()
	fi.Placeholder = "Filter results..."
	fi.CharLimit = 156
	fi.Width = 96
	fi.Prompt = " / "

	sp := spinner.New()
	sp.Spinner = spinner.Spinner{
		Frames: []string{"∙∙∙∙∙∙", "●∙∙∙∙∙", "∙●∙∙∙∙", "∙∙●∙∙∙", "∙∙∙●∙∙", "∙∙∙∙●∙", "∙∙∙∙∙●"},
		FPS:    time.Second / 10,
	}
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9"))

	columns := []table.Column{
		{Title: "URL", Width: 60},
		{Title: "Status", Width: 10},
		{Title: "Type", Width: 15},
		{Title: "      Size", Width: 10},
	}

	rows := []table.Row{}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(15),
	)
	t.KeyMap.LineUp.SetKeys("up", "k")
	t.KeyMap.LineDown.SetKeys("down", "j")

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false).
		Padding(0, 1)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("#bd93f9")).
		Bold(false)
	s.Cell = s.Cell.Padding(0, 1)
	t.SetStyles(s)

	return model{
		textInput:   ti,
		filterInput: fi,
		spinner:     sp,
		table:       t,
		visited:     make(map[string]bool),
		results:     make(chan crawlResult, 10000),
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
