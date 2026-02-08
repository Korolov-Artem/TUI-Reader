package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/taylorskalyo/goreader/epub"
)

func findEpubFiles() []string {
	files, err := os.ReadDir(".")
	if err != nil {
		return []string{}
	}

	var epubs []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".epub") {
			epubs = append(epubs, file.Name())
		}
	}
	return epubs
}

func updateLibrary(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.files) == 0 {
			return m, nil
		}

		selectedFile := m.files[m.cursor]
		rc, err := epub.OpenReader(selectedFile)
		if err != nil {
			return m, nil
		}

		m.currentBook = rc.Rootfiles[0]
		m.spine = m.currentBook.Spine.Itemrefs
		m.chapterIdx = 0
		m.pageIdx = 0

		m = loadChapter(m)
		m.state = readerView
		return m, nil
	}
	return m, nil
}

func viewLibrary(m model) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		PaddingLeft(2)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(4)

	s := "\n" + titleStyle.Render("MY LIBRARY") + "\n"

	if len(m.files) == 0 {
		s += "\n  No .epub files found in this directory.\n"
		s += "\n  (Press q to quit)"
		return s
	}

	for i, file := range m.files {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			s += fmt.Sprintf("%s%s\n", cursor, selectedStyle.Render(file))
		} else {
			s += fmt.Sprintf("%s%s\n", cursor, normalStyle.Render(file))
		}
	}

	s += "\n  [Use Arrows to Move | Enter to Select | q to Quit]\n"
	return s
}
