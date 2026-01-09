package main

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errorMsg error

type model struct {
	textInput   textinput.Model
	filterInput textinput.Model
	spinner     spinner.Model
	table       table.Model
	allRows     []table.Row
	visited     map[string]bool
	baseUrl     *url.URL
	width       int
	height      int
	crawler     *crawler
	results     chan crawlResult
	message     string
	msgTimer    *time.Timer
	crawling    bool
	finished    bool
	filtering   bool
	theme       Theme
}

type clearMsg struct{}
type crawlFinishedMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.waitForResults())
}

func (m model) waitForResults() tea.Cmd {
	return func() tea.Msg {
		return <-m.results
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update table width
		targetTableWidth := m.width - 3
		if targetTableWidth < 40 {
			targetTableWidth = 40
		}

		statusWidth := 10
		typeWidth := 15
		sizeWidth := 10
		urlWidth := targetTableWidth - statusWidth - typeWidth - sizeWidth - 3 - 8

		if urlWidth < 10 {
			urlWidth = 10
		}

		columns := []table.Column{
			{Title: "URL", Width: urlWidth},
			{Title: "Status", Width: statusWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "      Size", Width: sizeWidth},
		}
		m.table.SetColumns(columns)

		actualTableWidth := urlWidth + statusWidth + typeWidth + sizeWidth + 3 + 8 + 2
		leftInputWidth := actualTableWidth / 2
		rightInputWidth := actualTableWidth - leftInputWidth

		m.textInput.Width = leftInputWidth - 2
		m.filterInput.Width = rightInputWidth - 4

		tableHeight := m.height - 10
		m.table.SetHeight(tableHeight)

		return m, nil

	case clearMsg:
		m.message = ""
		return m, nil

	case crawlResult:
		if msg.url == "__FINISHED__" {
			m.crawling = false
			m.finished = true
			return m, nil
		}
		m.visited[msg.url] = true

		sizeKB := float64(msg.size) / 1024.0
		sizeStr := fmt.Sprintf("%.1f kB", sizeKB)
		formattedSize := fmt.Sprintf("%10s", sizeStr)

		row := table.Row{msg.url, msg.status, msg.kind, formattedSize}
		m.allRows = append(m.allRows, row)

		filter := strings.ToLower(m.filterInput.Value())
		if filter == "" || strings.Contains(strings.ToLower(msg.url), filter) {
			rows := m.table.Rows()
			rows = append(rows, row)
			m.table.SetRows(rows)
		}

		return m, m.waitForResults()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.crawler != nil {
				m.crawler.stop()
			}
			return m, tea.Quit
		case "q":
			if !m.filtering && !m.textInput.Focused() {
				if m.crawler != nil {
					m.crawler.stop()
				}
				return m, tea.Quit
			}
		case "esc":
			if m.filtering {
				m.filtering = false
				m.filterInput.Blur()
				m.table.SetRows(m.allRows)
				return m, nil
			}
			if m.crawler != nil {
				m.crawler.stop()
			}
			return m, tea.Quit
		case "tab":
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.filterInput.Focus()
				m.filtering = true
			} else if m.filterInput.Focused() {
				m.filterInput.Blur()
				m.table.Focus()
				m.filtering = false
			} else {
				m.table.Blur()
				m.textInput.Focus()
				m.filtering = false
			}
		case "shift+tab":
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.table.Focus()
				m.filtering = false
			} else if m.filterInput.Focused() {
				m.filterInput.Blur()
				m.textInput.Focus()
				m.filtering = false
			} else {
				m.table.Blur()
				m.filterInput.Focus()
				m.filtering = true
			}
		case "/":
			if m.table.Focused() && !m.filtering {
				m.filtering = true
				m.filterInput.Focus()
				m.filterInput.SetValue("")
				return m, nil
			}
		case "enter":
			if m.filtering || m.textInput.Focused() {
				if m.filtering {
					m.filtering = false
					m.filterInput.Blur()
				}
				rawUrl := m.textInput.Value()
				if rawUrl != "" {
					if !strings.HasPrefix(rawUrl, "http://") && !strings.HasPrefix(rawUrl, "https://") {
						rawUrl = "https://" + rawUrl
					}

					parsedUrl, err := url.Parse(rawUrl)
					if err != nil {
						return m, nil
					}

					if m.crawler != nil {
						m.crawler.stop()
						for len(m.results) > 0 {
							<-m.results
						}
					}

					m.baseUrl = parsedUrl
					m.visited = make(map[string]bool)
					m.allRows = []table.Row{}
					m.table.SetRows([]table.Row{})
					m.textInput.Blur()
					m.table.Focus()

					target := m.baseUrl.String()
					m.visited[target] = true

					concurrency := runtime.NumCPU() * 2
					if concurrency > maxConcurrency {
						concurrency = maxConcurrency
					}

					m.crawler = newCrawler(m.baseUrl, m.results, concurrency)
					m.crawler.visited.Store(target, true)
					m.crawler.active.Add(1)
					m.crawler.start()
					m.crawler.jobs <- target

					m.crawling = true
					m.finished = false

					return m, tea.Batch(
						m.waitForResults(),
						m.spinner.Tick,
					)
				}
				return m, nil
			} else if m.table.Focused() {
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					url := selectedRow[0]
					openURL(url)
				}
			}
		case "w":
			if m.table.Focused() {
				filename, err := m.exportToCSV()
				if err != nil {
					m.message = "Error exporting: " + err.Error()
				} else {
					m.message = "Exported: " + filename
				}
				return m, tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
					return clearMsg{}
				})
			}
		}
	}

	var tiCmd, fiCmd, tCmd tea.Cmd
	if m.textInput.Focused() {
		m.textInput, tiCmd = m.textInput.Update(msg)
	}
	if m.filterInput.Focused() {
		oldFilter := m.filterInput.Value()
		m.filterInput, fiCmd = m.filterInput.Update(msg)
		if m.filterInput.Value() != oldFilter {
			var filteredRows []table.Row
			filter := strings.ToLower(m.filterInput.Value())
			for _, row := range m.allRows {
				if strings.Contains(strings.ToLower(row[0]), filter) {
					filteredRows = append(filteredRows, row)
				}
			}
			m.table.SetRows(filteredRows)
		}
	}
	if m.table.Focused() {
		m.table, tCmd = m.table.Update(msg)
	}

	return m, tea.Batch(tiCmd, fiCmd, tCmd)
}

func (m model) View() string {
	var inputView, tableView string

	numResults := len(m.table.Rows())
	headerText := fmt.Sprintf(" Results: %d ", numResults)

	checkMarkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.CheckMarkColor))
	checkMark := checkMarkStyle.Render("✔")

	if m.crawling {
		headerText += fmt.Sprintf("• Crawling %s ", m.spinner.View())
	} else if m.finished {
		headerText += fmt.Sprintf("• Complete %s ", checkMark)
	}

	if m.message != "" {
		headerText += fmt.Sprintf("• %s ", m.message)
	}

	tableViewContent := m.table.View()
	contentWidth := lipgloss.Width(tableViewContent)
	if contentWidth == 0 {
		columns := m.table.Columns()
		for i, col := range columns {
			contentWidth += col.Width + 2
			if i < len(columns)-1 {
				contentWidth++
			}
		}
	}

	totalWidth := contentWidth + 2
	leftInputWidth := totalWidth / 2
	rightInputWidth := totalWidth - leftInputWidth

	// URL Input
	inputStyle := blurredStyle.Copy().Width(leftInputWidth - 2)
	if m.textInput.Focused() {
		inputStyle = focusedStyle.Copy().Width(leftInputWidth - 2)
	}
	inputStyle = inputStyle.Border(lipgloss.RoundedBorder()).BorderTop(false)
	inputView = inputStyle.Render(m.textInput.View())

	// Add intersecting title for URL Input
	inputTitle := " URL "
	if m.textInput.Focused() {
		inputTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.FocusedColor)).Render(inputTitle)
	}

	border := lipgloss.RoundedBorder()
	borderColor := m.theme.BlurredColor
	if m.textInput.Focused() {
		borderColor = m.theme.FocusedColor
	}
	bc := lipgloss.Color(borderColor)
	titleWidth := lipgloss.Width(inputTitle)

	left := lipgloss.NewStyle().Foreground(bc).Render(string(border.TopLeft) + string(border.Top))
	rightCount := leftInputWidth - titleWidth - 3
	if rightCount < 0 {
		rightCount = 0
	}
	right := lipgloss.NewStyle().Foreground(bc).Render(strings.Repeat(string(border.Top), rightCount) + string(border.TopRight))

	inputView = left + inputTitle + right + "\n" + inputView

	// Filter Input
	var filterView string
	filterStyle := blurredStyle.Copy().Width(rightInputWidth - 2)
	if m.filtering {
		filterStyle = focusedStyle.Copy().Width(rightInputWidth - 2)
	}
	filterStyle = filterStyle.Border(lipgloss.RoundedBorder()).BorderTop(false)
	filterView = filterStyle.Render(m.filterInput.View())

	// Add intersecting title for Filter Input
	filterTitle := " Filter "
	if m.filtering {
		filterTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.FocusedColor)).Render(filterTitle)
	}

	borderColor = m.theme.BlurredColor
	if m.filtering {
		borderColor = m.theme.FocusedColor
	}
	bc = lipgloss.Color(borderColor)
	titleWidth = lipgloss.Width(filterTitle)

	left = lipgloss.NewStyle().Foreground(bc).Render(string(border.TopLeft) + string(border.Top))
	rightCount = rightInputWidth - titleWidth - 3
	if rightCount < 0 {
		rightCount = 0
	}
	right = lipgloss.NewStyle().Foreground(bc).Render(strings.Repeat(string(border.Top), rightCount) + string(border.TopRight))

	filterView = left + filterTitle + right + "\n" + filterView

	inputsView := lipgloss.JoinHorizontal(lipgloss.Top, inputView, filterView)

	// Results Table
	baseTableStyle := blurredStyle.Copy().Width(contentWidth)
	if m.table.Focused() && !m.filtering {
		baseTableStyle = focusedStyle.Copy().Width(contentWidth)
	}
	baseTableStyle = baseTableStyle.Border(lipgloss.RoundedBorder()).BorderTop(false)
	tableView = baseTableStyle.Render(tableViewContent)

	// Add intersecting title for Results
	resultsTitle := headerText
	if m.table.Focused() && !m.filtering {
		resultsTitle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.FocusedColor)).Render(resultsTitle)
	}

	borderColor = m.theme.BlurredColor
	if m.table.Focused() && !m.filtering {
		borderColor = m.theme.FocusedColor
	}

	bc = lipgloss.Color(borderColor)
	titleWidth = lipgloss.Width(resultsTitle)

	left = lipgloss.NewStyle().Foreground(bc).Render(string(border.TopLeft) + string(border.Top))
	rightCount = (contentWidth + 2) - titleWidth - 3
	if rightCount < 0 {
		rightCount = 0
	}
	right = lipgloss.NewStyle().Foreground(bc).Render(strings.Repeat(string(border.Top), rightCount) + string(border.TopRight))

	tableView = left + resultsTitle + right + "\n" + tableView

	var helpView string
	if m.textInput.Focused() || m.filterInput.Focused() {
		helpView = "Tab: focus results • Enter: start crawl • Esc: quit"
	} else {
		helpView = "Tab: focus input • /: filter • Enter: open URL • w: export • Arrows/j/k: scroll • q: quit"
	}

	helpStyle := lipgloss.NewStyle().PaddingLeft(1)

	elements := []string{inputsView, tableView, helpStyle.Render(helpView)}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		elements...,
	)
}
