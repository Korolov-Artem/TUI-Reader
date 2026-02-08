package main

import (
	zip2 "archive/zip"

	"encoding/xml"
	"fmt"
	"io"

	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/taylorskalyo/goreader/epub"
)

func updateReader(m model, msg tea.KeyMsg) (model, tea.Cmd) {
	linesPerPage := m.height - 7
	if linesPerPage < 1 {
		linesPerPage = 1
	}

	allLines := strings.Split(m.currentText, "\n")
	totalLines := len(allLines)
	m.totalPages = (totalLines + linesPerPage - 1) / linesPerPage

	switch msg.String() {
	case "q", "esc":
		m.state = libraryView
		return m, nil
	case "n", "right":
		if m.pageIdx < m.totalPages-1 {
			m.pageIdx++
		} else {
			if m.chapterIdx < len(m.spine)-1 {
				m.chapterIdx++
				m.pageIdx = 0
				return loadChapter(m), nil
			}
		}
	case "p", "left":
		if m.pageIdx > 0 {
			m.pageIdx--
		} else {
			if m.chapterIdx > 0 {
				m.chapterIdx--
				m = loadChapter(m)

				lines := strings.Split(m.currentText, "\n")
				newTotal := (len(lines) + linesPerPage - 1) / linesPerPage
				m.pageIdx = newTotal - 1

				return m, nil
			}
		}
	}
	return m, nil
}

func viewReader(m model) string {
	textWidth := 80
	if m.width-10 < 80 {
		textWidth = m.width - 10
	}
	leftPadding := (m.width - textWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}
	pageStyle := lipgloss.NewStyle().Padding(1, 0, 1, leftPadding)
	textStyle := lipgloss.NewStyle().
		Width(textWidth).
		Align(lipgloss.Left)

	linesPerPage := m.height - 7
	if linesPerPage < 1 {
		linesPerPage = 1
	}

	wrappedText := textStyle.Render(m.currentText)

	allLines := strings.Split(wrappedText, "\n")
	totalLines := len(allLines)
	m.totalPages = (totalLines + linesPerPage - 1) / linesPerPage

	startLine := m.pageIdx * linesPerPage
	endLine := startLine + linesPerPage

	if startLine >= totalLines {
		startLine = 0
	}

	if endLine > totalLines {
		endLine = totalLines
	}

	pageLines := allLines[startLine:endLine]
	pageContent := strings.Join(pageLines, "\n")
	pageContent = strings.TrimPrefix(pageContent, "\n")

	view := pageStyle.Render(pageContent) + "\n"

	linesPrinted := endLine - startLine
	if linesPrinted < linesPerPage {
		view += strings.Repeat("\n", linesPerPage-linesPrinted)
	}

	status := fmt.Sprintf("Page %d/%d | Chapter %d/%d",
		m.pageIdx+1, m.totalPages, m.chapterIdx+1, len(m.spine))

	footerStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	view += footerStyle.Render(status) + "\n"
	view += footerStyle.Render("[ (n/Right) Next | (p/Left) Prev | (q) Library ]")

	return view
}

func loadChapter(m model) model {
	item := m.spine[m.chapterIdx]

	fileName := findFileName(m.currentBook, item.ID)
	if fileName == "" {
		m.currentText = "[Error: Chapter file not found]"
		return m
	}

	zip, err := zip2.OpenReader(m.files[m.cursor])
	if err != nil {
		m.currentText = "[Error: Could not open epub zip]"
		return m
	}
	defer zip.Close()

	fileStream, err := openZipFile(zip, fileName)
	if err != nil {
		m.currentText = "[Error: Could not open chapter file]"
		return m
	}

	rawText := extractText(fileStream)
	fileStream.Close()

	m.currentText = strings.ReplaceAll(rawText, "\t", "    ")
	m.totalPages = 1
	return m
}

func findFileName(book *epub.Rootfile, chapterId string) string {
	for _, manifestItem := range book.Manifest.Items {
		if manifestItem.ID == chapterId {
			return manifestItem.HREF
		}
	}
	return ""
}

func openZipFile(zipReader *zip2.ReadCloser, targetFileName string) (io.ReadCloser, error) {
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, targetFileName) {
			return file.Open()
		}
	}
	return nil, fmt.Errorf("file not found: %s", targetFileName)
}

func extractText(readableData io.Reader) string {
	decoder := xml.NewDecoder(readableData)
	var sb strings.Builder

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}

		switch token := t.(type) {
		case xml.StartElement:
			if token.Name.Local == "p" || token.Name.Local == "div" ||
				token.Name.Local == "br" || token.Name.Local == "h1" ||
				token.Name.Local == "h2" || token.Name.Local == "li" {
				sb.WriteString("\n\n")
			}
		case xml.CharData:
			content := string(token)

			sb.WriteString(content)
		}
	}
	return sb.String()
}
