package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/taylorskalyo/goreader/epub"
)

type sessionState int

const (
	libraryView sessionState = iota
	readerView
)

type model struct {
	state sessionState

	files  []string
	cursor int

	currentBook *epub.Rootfile
	spine       []epub.Itemref
	chapterIdx  int
	pageIdx     int

	currentText string
	totalPages  int

	width  int
	height int
}

func main() {
	initialModel := model{
		state: libraryView,
		files: findEpubFiles(),
	}

	program := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Printf("There's been an error: %v", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "q" && m.state == libraryView {
			return m, tea.Quit
		}

		if m.state == libraryView {
			return updateLibrary(m, msg)
		} else {
			return updateReader(m, msg)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.state == libraryView {
		return viewLibrary(m)
	} else {
		return viewReader(m)
	}
}
