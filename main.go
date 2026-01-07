package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

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
)

type crawlResult struct {
	url    string
	status string
	kind   string
	size   int64
	links  []string
}

type errorMsg error

type model struct {
	textInput textinput.Model
	table     table.Model
	visited   map[string]bool
	baseUrl   *url.URL
}

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
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return model{
		textInput: ti,
		table:     t,
		visited:   make(map[string]bool),
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func crawl(targetUrl string, baseUrl *url.URL) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(targetUrl)
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
			links = extractLinks(strings.NewReader(string(bodyBytes)), targetUrl, baseUrl)
		}

		return crawlResult{
			url:    targetUrl,
			status: fmt.Sprintf("%d", resp.StatusCode),
			kind:   kind,
			size:   int64(len(bodyBytes)),
			links:  links,
		}
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
	case crawlResult:
		// Mark as visited
		m.visited[msg.url] = true

		// Add to table
		rows := m.table.Rows()
		rows = append(rows, table.Row{msg.url, msg.status, msg.kind, fmt.Sprintf("%d", msg.size)})
		m.table.SetRows(rows)

		// Find new links to crawl
		var cmds []tea.Cmd
		for _, link := range msg.links {
			if !m.visited[link] {
				m.visited[link] = true // Mark as visited immediately to avoid duplicate queues
				cmds = append(cmds, crawl(link, m.baseUrl))
			}
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
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

					m.baseUrl = parsedUrl
					m.visited = make(map[string]bool)
					m.table.SetRows([]table.Row{})
					m.textInput.Blur()
					m.table.Focus()

					target := m.baseUrl.String()
					m.visited[target] = true
					return m, crawl(target, m.baseUrl)
				}
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

	if m.textInput.Focused() {
		inputView = focusedStyle.Render(m.textInput.View())
	} else {
		inputView = blurredStyle.Render(m.textInput.View())
	}

	numResults := len(m.table.Rows())
	headerText := fmt.Sprintf(" Results: %d", numResults)

	// Create a style for the header to match the table width
	headerStyle := lipgloss.NewStyle().
		Width(98).
		Border(lipgloss.RoundedBorder(), true, true, false, true)

	if m.table.Focused() {
		headerStyle = headerStyle.BorderForeground(lipgloss.Color("#bd93f9"))
		// Adjust table style to remove top border since header provides it
		tableView = focusedStyle.Copy().
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color("#bd93f9")).
			Render(m.table.View())
	} else {
		headerStyle = headerStyle.BorderForeground(lipgloss.Color("240"))
		// Adjust table style to remove top border since header provides it
		tableView = blurredStyle.Copy().
			Border(lipgloss.RoundedBorder(), false, true, true, true).
			BorderForeground(lipgloss.Color("240")).
			Render(m.table.View())
	}

	headerView := headerStyle.Render(headerText)

	return fmt.Sprintf(
		"%s\n%s\n%s\n\n%s",
		inputView,
		headerView,
		tableView,
		"Tab: switch focus • Enter: start crawl • Arrows: scroll • q: quit",
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
