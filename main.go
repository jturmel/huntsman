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
	concurrency int
}

func newCrawler(baseUrl *url.URL, results chan crawlResult, concurrency int) *crawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &crawler{
		baseUrl:     baseUrl,
		results:     results,
		jobs:        make(chan string, 1000),
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
				return
			case c.results <- res:
			}

			for _, link := range res.links {
				if _, loaded := c.visited.LoadOrStore(link, true); !loaded {
					select {
					case <-c.ctx.Done():
						return
					case c.jobs <- link:
					default:
						// If jobs channel is full, we might want to log or handle it
						// For now, we'll just skip to avoid deadlocks in this simple implementation
					}
				}
			}
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
		kind = "HTML"
	}

	var links []string
	if kind == "HTML" {
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
	textInput textinput.Model
	table     table.Model
	visited   map[string]bool
	baseUrl   *url.URL
	width     int
	height    int
	crawler   *crawler
	results   chan crawlResult
	message   string
	msgTimer  *time.Timer
}

type clearMsg struct{}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL to spider..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 96
	ti.Prompt = " "

	columns := []table.Column{
		{Title: "URL", Width: 60},
		{Title: "Status", Width: 10},
		{Title: "Type", Width: 10},
		{Title: "Size", Width: 10},
	}

	rows := []table.Row{}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(15),
	)

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
		Bold(false).
		Padding(0, 1)
	s.Cell = s.Cell.Padding(0, 1)
	t.SetStyles(s)

	return model{
		textInput: ti,
		table:     t,
		visited:   make(map[string]bool),
		results:   make(chan crawlResult, 100),
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
		typeWidth := 10
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
			{Title: "Size", Width: sizeWidth},
		}
		m.table.SetColumns(columns)

		// Adjust input width to match the table's total width
		// Table total width = sum(column widths) + spaces between columns + padding
		actualTableWidth := urlWidth + statusWidth + typeWidth + sizeWidth + 3 + 8
		m.textInput.Width = actualTableWidth - 1 // -1 for prompt space

		// Adjust table height
		// Total height minus input (3 lines), header (3 lines), help (1 line) and some buffer
		m.table.SetHeight(m.height - 10)

		return m, nil

	case clearMsg:
		m.message = ""
		return m, nil

	case crawlResult:
		// Mark as visited in UI
		m.visited[msg.url] = true

		// Add to table
		rows := m.table.Rows()
		rows = append(rows, table.Row{msg.url, msg.status, msg.kind, fmt.Sprintf("%d", msg.size)})
		m.table.SetRows(rows)

		return m, m.waitForResults()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
			if m.crawler != nil {
				m.crawler.stop()
			}
			return m, tea.Quit
		case "tab", "shift+tab":
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.table.Focus()
			} else {
				m.textInput.Focus()
				m.table.Blur()
			}
		case "enter":
			if m.textInput.Focused() {
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
					m.crawler.start()
					m.crawler.jobs <- target

					return m, m.waitForResults()
				}
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

	var tiCmd, tCmd tea.Cmd
	if m.textInput.Focused() {
		m.textInput, tiCmd = m.textInput.Update(msg)
	}
	if m.table.Focused() {
		m.table, tCmd = m.table.Update(msg)
	}

	return m, tea.Batch(tiCmd, tCmd)
}

func (m model) View() string {
	var inputView, tableView string

	numResults := len(m.table.Rows())
	headerText := fmt.Sprintf(" Results: %d", numResults)
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

	if m.table.Focused() {
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
	inputStyle := blurredStyle.Copy().
		Width(contentWidth)
	if m.textInput.Focused() {
		inputStyle = focusedStyle.Copy().
			Width(contentWidth)
	}
	inputView = inputStyle.Render(m.textInput.View())

	var helpView string
	if m.textInput.Focused() {
		helpView = "Tab: focus table • Enter: start crawl • q: quit"
	} else {
		helpView = "Tab: focus input • Enter: open URL • w: export • Arrows/j/k: scroll • q: quit"
	}

	helpStyle := lipgloss.NewStyle().PaddingLeft(1)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		headerView,
		tableView,
		helpStyle.Render(helpView),
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
