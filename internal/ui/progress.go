package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
)

const (
	padding  = 2
	maxWidth = 80
)

type model struct {
	progress     progress.Model
	current      int
	total        int
	latestDetail ComicDetail
}

type ComicDetail struct {
	Title       string
	IssueNumber int
	Publisher   string
	ReleaseDate time.Time
	CoverPrice  int64
	Value       int64
	ImageUrl    string
	UPC         string
	Box         string
}

type ComicScrapeMessage struct {
	Current int
	Detail  ComicDetail
}

func NewProgressModel(total int) model {
	return model{
		progress: progress.New(progress.WithDefaultBlend()),
		current:  0,
		total:    total,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ComicScrapeMessage:
		m.current = msg.Current
		m.latestDetail = msg.Detail
		cmd := m.progress.SetPercent(float64(m.current) / float64(m.total))
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
	v := tea.NewView("\n" +
		pad + "Scraping comics.... " +
		m.progress.View() +
		fmt.Sprintf(" %d/%d", m.current, m.total) +
		"\n\n" +
		pad + fmt.Sprintf("Title:\t\t\t\t%s\n", m.latestDetail.Title) +
		pad + fmt.Sprintf("Issue Number:\t\t\t%d\n", m.latestDetail.IssueNumber) +
		pad + fmt.Sprintf("Publisher:\t\t\t%s\n", m.latestDetail.Publisher) +
		pad + fmt.Sprintf("ReleaseDate:\t\t\t%s\n", m.latestDetail.ReleaseDate.Format("02. Jan 2006")) +
		pad + fmt.Sprintf("Cover Price:\t\t\t$%.2f\n", float64(m.latestDetail.CoverPrice)/100.0) +
		pad + fmt.Sprintf("Value:\t\t\t\t$%.2f\n", float64(m.latestDetail.Value)/100.0) +
		pad + fmt.Sprintf("UPC:\t\t\t\t%s\n", m.latestDetail.UPC) +
		pad + fmt.Sprintf("Box:\t\t\t\t%s\n", m.latestDetail.Box) +
		"\n\n")
	v.AltScreen = true
	return v
}
