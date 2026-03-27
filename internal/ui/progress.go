package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	padding  = 2
	maxWidth = 80
)

var asciiLogo = `
     ████████████████████████      ██████████████████████       ███████████████████████
    ██████████████████████████   ████████████████████████████  ██████████████████████████
    ████                         ████                  █████   ████
    ████                         ████                   ████   ████
    ████                         █████████████████████         ████
    ████                         ███████████████████████████   ████
    ████                         ████                   █████  ████
    ████                         ████                    ████  ████
    ██████████████████████████   ████████████████████████████  ██████████████████████████
     ████████████████████████      ██████████████████████       ███████████████████████  `

func header() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 4).
		Foreground(lipgloss.Color("63")).
		Width(maxWidth + 25)

	return borderStyle.Render(asciiLogo + "\n    Comic Book Collector")
}

func showComicDetail(detail ComicDetail) string {
	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 6).
		Foreground(lipgloss.White).
		Width(maxWidth + 25)

	pad := strings.Repeat(" ", padding)

	return borderStyle.Render(
		"\n\n" +
			pad + fmt.Sprintf("%-20s %s\n", "Title:", detail.Title) +
			pad + fmt.Sprintf("%-20s %d\n", "Issue Number:", detail.IssueNumber) +
			pad + fmt.Sprintf("%-20s %s\n", "Publisher:", detail.Publisher) +
			pad + fmt.Sprintf("%-20s %s\n", "ReleaseDate:", detail.ReleaseDate.Format("02. Jan 2006")) +
			pad + fmt.Sprintf("%-20s %.2f\n", "Cover Price:", float64(detail.CoverPrice)/100.0) +
			pad + fmt.Sprintf("%-20s %.2f\n", "Value:", float64(detail.Value)/100.0) +
			pad + fmt.Sprintf("%-20s %s\n", "UPC:", detail.UPC) +
			pad + fmt.Sprintf("%-20s %s\n", "Box:", detail.Box) +
			"\n\n")
}

type model struct {
	progress     progress.Model
	current      int
	total        int
	status       string
	latestDetail ComicDetail
}

type ComicDetail struct {
	Title       string
	IssueNumber int
	Publisher   string
	ReleaseDate time.Time
	CoverPrice  int64
	Value       int64
	UPC         string
	Box         string
}

type ComicScrapeMessage struct {
	Total         int
	Current       int
	StatusMessage string
	Detail        ComicDetail
}

func NewProgressModel() model {
	return model{
		progress: progress.New(progress.WithDefaultBlend()),
		current:  0,
		total:    0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ComicScrapeMessage:
		m.total = msg.Total
		m.current = msg.Current
		m.status = msg.StatusMessage
		m.latestDetail = msg.Detail
		var cmd tea.Cmd
		if m.total > 0 {
			cmd = m.progress.SetPercent(float64(m.current) / float64(m.total))
		}
		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "Q":
			return m, tea.Quit
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.progress.SetWidth(msg.Width - padding*2 - 4)
		if m.progress.Width() > maxWidth {
			m.progress.SetWidth(maxWidth)
		}
		return m, nil
	case progress.FrameMsg:
		// Update the progress bar frame
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd
	default:
		// Handle other messages if needed
		return m, nil
	}
}

func (m model) View() tea.View {
	pad := strings.Repeat(" ", padding)
	var content string
	content = header() + "\n\n"
	if m.total == 0 {
		content += "\n" + pad + m.status + "\n\n"
	} else if m.current == 0 {
		content += "\n" +
			pad + m.status +
			m.progress.View() +
			fmt.Sprintf(" %d/%d", m.current, m.total)
	} else {
		content += "\n" +
			pad + m.status +
			m.progress.View() +
			fmt.Sprintf(" %d/%d", m.current, m.total) +
			"\n\n\n" +
			showComicDetail(m.latestDetail) +
			"\n\n"
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
