package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	contentLines   []string
	wrappedPages   [][]string
	currentPage    int
	totalPages     int
	terminalWidth  int
	terminalHeight int
	message        string
	contentStyle   lipgloss.Style
	statusStyle    lipgloss.Style
	helpStyle      lipgloss.Style
}

const bookFilePath = "book.txt"

func loadBookCmd() tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(bookFilePath)
		if err != nil {
			return errMsg{err}
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return errMsg{err}
		}
		return bookLoadedMsg(lines)
	}
}

type bookLoadedMsg []string
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func (m *model) Init() tea.Cmd {
	m.contentStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9e9e9e")).
		Background(lipgloss.Color("#121212")).
		PaddingTop(1).
		Align(lipgloss.Left)

	m.statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9e9e9e")).
		Background(lipgloss.Color("#121212")).
		Align(lipgloss.Center).
		Width(m.terminalWidth).
		MarginLeft(1)

	m.helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#747474")).
		Background(lipgloss.Color("#121212")).
		Align(lipgloss.Left).
		Padding(0, 2)

	return tea.Batch(
		loadBookCmd(),
		tea.WindowSize(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "pgup", "right", "l", " ":
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
			}
		case "pgdown", "left", "h":
			if m.currentPage > 0 {
				m.currentPage--
			}
		case "home":
			m.currentPage = 0
		case "end":
			m.currentPage = m.totalPages - 1
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		m.rePaginateContent()
	case bookLoadedMsg:
		m.contentLines = msg
		m.rePaginateContent()
	case errMsg:
		m.message = fmt.Sprintf("Error: %s\n", msg.Error())
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) View() string {
	if m.contentLines == nil || len(m.contentLines) == 0 ||
		m.wrappedPages == nil || len(m.wrappedPages) == 0 {
		if m.message != "" {
			return m.message
		}
		return "Loading book...\n\nPress q to quit"
	}
	contentAreaWidth := m.terminalWidth - m.contentStyle.GetHorizontalFrameSize()
	contentAreaHeight := m.terminalHeight - m.contentStyle.GetVerticalFrameSize() -
		m.statusStyle.GetVerticalFrameSize()

	if contentAreaWidth < 1 {
		contentAreaWidth = 1
	}

	var currentPageContent []string
	if m.currentPage < len(m.wrappedPages) {
		currentPageContent = m.wrappedPages[m.currentPage]
	} else {
		currentPageContent = []string{"Error: Page not found."}
	}

	pageString := strings.Join(currentPageContent, "\n")
	renderedContent := m.contentStyle.Copy().
		Width(contentAreaWidth).
		Height(contentAreaHeight).
		Render(pageString)

	statusText := fmt.Sprintf("Page %d/%d", m.currentPage+1, m.totalPages)
	renderedStatus := m.statusStyle.Copy().
		Width(m.terminalWidth).
		Render(statusText)

	helpText := fmt.Sprintf("Press PgUp/PgDown or H/L or Left/Right to navigate," +
		" Home/End to go to start or end, Q or Ctrl+C to quit.")
	renderedHelp := m.helpStyle.Copy().
		Width(m.terminalWidth).
		Align(lipgloss.Center).
		Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderedContent,
		renderedStatus,
		renderedHelp,
	)
}

func (m *model) rePaginateContent() {
	if m.terminalWidth == 0 || m.terminalHeight == 0 || len(m.contentLines) == 0 {
		return
	}
	contentAvailableWidth := m.terminalWidth - m.contentStyle.GetHorizontalFrameSize()
	if contentAvailableWidth < 1 {
		contentAvailableWidth = 1
	}

	contentAvailableHeight := m.terminalHeight - m.contentStyle.GetVerticalFrameSize() -
		m.statusStyle.GetVerticalFrameSize() - m.helpStyle.GetVerticalFrameSize() - 2
	if contentAvailableHeight < 1 {
		contentAvailableHeight = 1
	}

	m.wrappedPages = [][]string{}
	currentPageVisualLines := []string{}
	currentVisualHeight := 0

	tempWrapperStyle := lipgloss.NewStyle().Width(contentAvailableWidth)

	for _, line := range m.contentLines {
		wrappedLine := tempWrapperStyle.Render(line)
		visualLines := strings.Split(wrappedLine, "\n")

		for _, vLine := range visualLines {
			if currentVisualHeight+1 > contentAvailableHeight {
				m.wrappedPages = append(m.wrappedPages, currentPageVisualLines)
				currentPageVisualLines = []string{}
				currentVisualHeight = 0
			}
			currentPageVisualLines = append(currentPageVisualLines, vLine)
			currentVisualHeight++
		}
	}

	if len(currentPageVisualLines) > 0 {
		m.wrappedPages = append(m.wrappedPages, currentPageVisualLines)
	}

	m.totalPages = len(m.wrappedPages)

	if m.currentPage >= m.totalPages {
		m.currentPage = m.totalPages - 1
	}
	if m.currentPage < 0 {
		m.currentPage = 0
	}
}

func main() {
	initialModel := &model{
		currentPage: 0,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
