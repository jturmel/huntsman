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
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
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

type crawlResult struct {
	url    string
	status string
	kind   string
	size   int64
	links  []string
}

type crawler struct {
	baseUrl     *url.URL
	visited     sync.Map
	results     chan crawlResult
	jobs        chan string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	active      sync.WaitGroup
	concurrency int
}

func newCrawler(baseUrl *url.URL, results chan crawlResult, concurrency int) *crawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &crawler{
		baseUrl:     baseUrl,
		results:     results,
		jobs:        make(chan string, 10000),
		ctx:         ctx,
		cancel:      cancel,
		concurrency: concurrency,
	}
}

func (c *crawler) start() {
	for i := 0; i < c.concurrency; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	// Monitor for completion
	go func() {
		c.active.Wait()
		// No more active jobs, close jobs to signal workers to stop if they are waiting for jobs
		// Actually, we should probably send a signal that we are done.
		// Since we don't want to close results channel (the model might still be processing it),
		// we can use a special result or another way to signal.
		// Given the current structure, let's just send a nil-like result or a special type if possible.
		// But wait, crawlResult is a struct.
		c.results <- crawlResult{url: "__FINISHED__"}
	}()
}

func (c *crawler) stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *crawler) worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case target, ok := <-c.jobs:
			if !ok {
				return
			}
			res := c.doCrawl(target)

			select {
			case <-c.ctx.Done():
				c.active.Done()
				return
			case c.results <- res:
			}

			for _, link := range res.links {
				if _, loaded := c.visited.LoadOrStore(link, true); !loaded {
					c.active.Add(1)
					select {
					case <-c.ctx.Done():
						c.active.Done()
						return
					case c.jobs <- link:
					default:
						c.active.Done()
					}
				}
			}
			c.active.Done()
		}
	}
}

func (c *crawler) doCrawl(targetUrl string) crawlResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(targetUrl)
	if err != nil {
		return crawlResult{url: targetUrl, status: "Error", kind: "N/A", size: 0}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return crawlResult{url: targetUrl, status: "Read Err", kind: "N/A", size: 0}
	}

	contentType := resp.Header.Get("Content-Type")
	kind := "Other"
	if strings.Contains(contentType, "text/html") {
		kind = "document"
	} else if strings.Contains(contentType, "text/css") {
		kind = "stylesheet"
	} else if strings.Contains(contentType, "javascript") {
		kind = "script"
	} else if strings.Contains(contentType, "font") {
		kind = "font"
	} else if strings.Contains(contentType, "image/png") {
		kind = "png"
	} else if strings.Contains(contentType, "image/gif") {
		kind = "gif"
	} else if strings.Contains(contentType, "image/jpeg") {
		kind = "jpeg"
	} else if strings.Contains(contentType, "image/svg+xml") {
		kind = "svg+xml"
	} else if strings.Contains(contentType, "x-icon") || strings.Contains(contentType, "vnd.microsoft.icon") {
		kind = "x-icon"
	} else if strings.Contains(contentType, "manifest+json") {
		kind = "manifest"
	}

	var links []string
	if kind == "document" {
		links = extractLinks(strings.NewReader(string(bodyBytes)), targetUrl, c.baseUrl)
	}

	return crawlResult{
		url:    targetUrl,
		status: fmt.Sprintf("%d", resp.StatusCode),
		kind:   kind,
		size:   int64(len(bodyBytes)),
		links:  links,
	}
}

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
}

type clearMsg struct{}
type crawlFinishedMsg struct{}

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

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.waitForResults())
}

func (m model) waitForResults() tea.Cmd {
	return func() tea.Msg {
		return <-m.results
	}
}

func extractLinks(body io.Reader, currentUrl string, baseUrl *url.URL) []string {
	var links []string
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			var attrKey string
			switch t.Data {
			case "a", "link":
				attrKey = "href"
			case "img", "script", "video", "audio", "source":
				attrKey = "src"
			default:
				continue
			}

			for _, a := range t.Attr {
				if a.Key == attrKey {
					u, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					resolved := baseUrl.ResolveReference(u)
					// Only include links on the same domain
					if resolved.Host == baseUrl.Host {
						// Strip fragment for normalization
						resolved.Fragment = ""
						links = append(links, resolved.String())
					}
				}
			}
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update table width
		// We subtract 3 characters for the borders and a safety margin to prevent wrapping
		// totalWidth = columnsWidths + (numColumns - 1)
		// We want totalWidth to be around m.width - 3
		targetTableWidth := m.width - 3
		if targetTableWidth < 40 {
			targetTableWidth = 40
		}

		// Adjust columns proportionally
		// URL gets the most space
		statusWidth := 10
		typeWidth := 15
		sizeWidth := 10
		// bubbles/table adds 1 space between columns.
		// We have 4 columns, so 3 spaces.
		// Additionally, each column now has 1 space padding on left and right (total 2 per column).
		// 4 columns * 2 padding = 8 spaces of padding.
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

		// Adjust input width to match the table's total width
		// Table total width = sum(column widths) + spaces between columns + padding + 2 for outer borders
		actualTableWidth := urlWidth + statusWidth + typeWidth + sizeWidth + 3 + 8 + 2
		leftInputWidth := actualTableWidth / 2
		rightInputWidth := actualTableWidth - leftInputWidth

		m.textInput.Width = leftInputWidth - 2    // -2 for borders
		m.filterInput.Width = rightInputWidth - 4 // -4 for borders and " / " prompt

		// Adjust table height
		// Total height minus input (3 lines), header (3 lines), help (1 line) and some buffer
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
		// Mark as visited in UI
		m.visited[msg.url] = true

		// Add to allRows
		sizeKB := float64(msg.size) / 1024.0
		sizeStr := fmt.Sprintf("%.1f kB", sizeKB)
		formattedSize := fmt.Sprintf("%10s", sizeStr)

		row := table.Row{msg.url, msg.status, msg.kind, formattedSize}
		m.allRows = append(m.allRows, row)

		// Filter and update table if it matches filter or if filter is empty
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
						// Could handle error here, for now just ignore
						return m, nil
					}

					if m.crawler != nil {
						m.crawler.stop()
						// Drain any remaining results from the old crawler
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
			// Update table rows based on filter
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
	headerText := fmt.Sprintf(" Results: %d", numResults)

	checkMarkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#bd93f9"))
	checkMark := checkMarkStyle.Render("✔")

	if m.crawling {
		headerText += fmt.Sprintf(" • Crawling %s", m.spinner.View())
	} else if m.finished {
		headerText += fmt.Sprintf(" • Complete %s", checkMark)
	}

	if m.message != "" {
		headerText += fmt.Sprintf(" • %s", m.message)
	}

	// Use the actual rendered width of the table to ensure alignment
	tableViewContent := m.table.View()
	contentWidth := lipgloss.Width(tableViewContent)
	if contentWidth == 0 {
		// Fallback to manual calculation if table is not yet rendered or has no columns
		columns := m.table.Columns()
		for i, col := range columns {
			contentWidth += col.Width + 2 // width + left/right padding
			if i < len(columns)-1 {
				contentWidth++ // space between columns
			}
		}
	}

	headerStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Border(lipgloss.RoundedBorder(), true, true, false, true)

	if m.table.Focused() && !m.filtering {
		headerStyle = headerStyle.BorderForeground(lipgloss.Color("#bd93f9"))
		// Adjust table style to remove top border since header provides it
		tableView = focusedStyle.Copy().
			Width(contentWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color("#bd93f9")).
			Render(tableViewContent)
	} else {
		headerStyle = headerStyle.BorderForeground(lipgloss.Color("240"))
		// Adjust table style to remove top border since header provides it
		tableView = blurredStyle.Copy().
			Width(contentWidth).
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color("240")).
			Render(tableViewContent)
	}

	headerView := headerStyle.Render(headerText)

	// Ensure input view also respects width
	// Total width should be contentWidth + 2 to account for the table's outer borders
	totalWidth := contentWidth + 2
	leftInputWidth := totalWidth / 2
	rightInputWidth := totalWidth - leftInputWidth

	inputStyle := blurredStyle.Copy().
		Width(leftInputWidth - 2) // -2 for borders
	if m.textInput.Focused() {
		inputStyle = focusedStyle.Copy().
			Width(leftInputWidth - 2)
	}
	inputView = inputStyle.Render(m.textInput.View())

	filterStyle := blurredStyle.Copy().
		Width(rightInputWidth - 2)
	if m.filtering {
		filterStyle = focusedStyle.Copy().
			Width(rightInputWidth - 2)
	}
	filterView := filterStyle.Render(m.filterInput.View())

	inputsView := lipgloss.JoinHorizontal(lipgloss.Top, inputView, filterView)

	var helpView string
	if m.textInput.Focused() || m.filterInput.Focused() {
		helpView = "Tab: focus results • Enter: start crawl • Esc: quit"
	} else {
		helpView = "Tab: focus input • /: filter • Enter: open URL • w: export • Arrows/j/k: scroll • q: quit"
	}

	helpStyle := lipgloss.NewStyle().PaddingLeft(1)

	elements := []string{inputsView, headerView, tableView, helpStyle.Render(helpView)}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		elements...,
	)
}

func (m model) exportToCSV() (string, error) {
	if m.baseUrl == nil {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	downloadsDir := filepath.Join(home, "Downloads")
	// If Downloads doesn't exist (unlikely but possible on some minimal setups), fallback to Home
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

	// Header
	_ = writer.Write([]string{"URL", "Status", "Type", "Size"})

	for _, row := range m.table.Rows() {
		// The size is already formatted as "%.2f kB" in the table row
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
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{u}
	}
	_ = exec.Command(cmd, args...).Start()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
