package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	contentLines   []string
	currentPage    int
	totalPages     int
	terminalWidth  int
	terminalHeight int
	linesPerPage   int
	message        string
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
	return tea.Batch(
		loadBookCmd(),
		tea.WindowSize(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
		m.calculatePagination()
	case bookLoadedMsg:
		m.contentLines = msg
		m.calculatePagination()
	case errMsg:
		m.message = fmt.Sprintf("Error: %s\n", msg.Error())
		return m, tea.Quit
	}
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.contentLines == nil || len(m.contentLines) == 0 {
		if m.message != "" {
			return m.message
		}
		return "Loading book..."
	}
	startLine := m.currentPage * m.linesPerPage
	endLine := startLine + m.linesPerPage
	if endLine > len(m.contentLines) {
		endLine = len(m.contentLines)
	}

	pageContent := m.contentLines[startLine:endLine]
	pageString := strings.Join(pageContent, "\n")

	statusLine := fmt.Sprintf("Page %d/%d (Width: %d, Height: %d)",
		m.currentPage+1, m.totalPages, m.terminalWidth, m.terminalHeight)

	fillerLines := m.linesPerPage - len(pageContent)
	if fillerLines < 0 {
		fillerLines = 0
	}
	for i := 0; i < fillerLines; i++ {
		pageString += "\n"
	}
	return fmt.Sprintf("%s\n\n%s\n\n"+
		"Press PgUp/PgDown or H/L or Left/Right to navigate,"+
		" Home/End to go to start or end, Q or Ctrl+C to quit.", pageString, statusLine)
}

func (m *model) calculatePagination() {
	if m.terminalHeight == 0 || m.terminalWidth == 0 {
		return
	}
	reservedLines := 5
	if m.terminalWidth > reservedLines {
		m.linesPerPage = m.terminalHeight - reservedLines
	} else {
		m.linesPerPage = 1
	}
	if m.linesPerPage <= 0 {
		m.linesPerPage = 1
	}

	m.totalPages = (len(m.contentLines) + m.linesPerPage - 1) / m.linesPerPage

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
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
