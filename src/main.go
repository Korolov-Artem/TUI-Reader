package main

import (
	zip2 "archive/zip"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/taylorskalyo/goreader/epub"
	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide an epub filename")
		return
	}
	filename := os.Args[1]

	rc, err := epub.OpenReader(filename)
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	zip, err := zip2.OpenReader(filename)
	if err != nil {
		panic(err)
	}
	defer zip.Close()

	book := rc.Rootfiles[0]
	fmt.Println(book.Title)

	spine := book.Spine.Itemrefs
	currentIndex := 0

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\033[H\033[2J")

		item := spine[currentIndex]

		fmt.Printf("Chapter %d of %d (ID: %s)\n", currentIndex+1, len(spine), item.ID)

		docStyle, textStyle, totalHeight, totalWidth := createReaderStyles()

		fileName := findFileName(book, item.ID)

		if fileName == "" {
			handleEmptyText(docStyle, reader, &currentIndex, spine)
			continue
		}

		fileStream, err := openZipFile(zip, fileName)
		if err != nil {
			handleNavigation(reader, &currentIndex, len(spine))
			continue
		}

		rawText := extractText(fileStream)
		fileStream.Close()
		formattedText := strings.ReplaceAll(rawText, "\t", "    ")

		if strings.TrimSpace(rawText) == "" {
			handleEmptyText(docStyle, reader, &currentIndex, spine)
			continue
		} else {
			viewChapter(totalHeight, totalWidth, textStyle, docStyle, formattedText, reader, &currentIndex, spine)
		}

		if currentIndex >= len(spine) {
			currentIndex = len(spine) - 1
		}
		if currentIndex < 0 {
			currentIndex = 0
		}
	}

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

func handleNavigation(reader *bufio.Reader, currentIndex *int, totalChapters int) {
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "n":
		*currentIndex++
	case "p":
		*currentIndex--
	case "q":
		os.Exit(0)
	}

	if *currentIndex >= totalChapters {
		*currentIndex = totalChapters - 1
	}
	if *currentIndex < 0 {
		*currentIndex = 0
	}
}

func handleEmptyText(docStyle lipgloss.Style, reader *bufio.Reader, currentIndex *int, spine []epub.Itemref) {
	fmt.Println("\n[This chapter contains no text (Image or Empty)]")
	fmt.Println(docStyle.Render("\n[ Press (n) Next or (p) Prev ]"))
	handleNavigation(reader, currentIndex, len(spine))
}

func createReaderStyles() (docStyle, textStyle lipgloss.Style, totalHeight, totalWidth int) {
	totalWidth, totalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || totalWidth <= 0 {
		totalWidth = 80
		totalHeight = 64
	}

	docStyle = lipgloss.NewStyle().
		Width(totalWidth).
		Align(lipgloss.Center)

	textStyle = lipgloss.NewStyle().
		Width(min(80, totalWidth-10)).
		Align(lipgloss.Left)

	return docStyle, textStyle, totalHeight, totalWidth
}

func viewChapter(totalHeight, totalWidth int, textStyle, docStyle lipgloss.Style, formattedText string, reader *bufio.Reader, currentIndex *int, spine []epub.Itemref) {
	verticalMargin := 7
	linesPerPage := totalHeight - verticalMargin
	if linesPerPage < 1 {
		linesPerPage = 1
	}

	textWidth := 80
	if totalWidth-10 < 80 {
		textWidth = totalWidth - 10
	}
	leftPadding := (totalWidth - textWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	pageStyle := lipgloss.NewStyle().Padding(1, 0, 1, leftPadding)

	fullFormattedText := textStyle.Render(formattedText)
	allLines := strings.Split(fullFormattedText, "\n")

	totalLines := len(allLines)
	totalPages := (totalLines + linesPerPage - 1) / linesPerPage

	currentPage := 0

	for {
		startLine := currentPage * linesPerPage
		endLine := startLine + linesPerPage
		if endLine > totalLines {
			endLine = totalLines
		}

		pageLines := allLines[startLine:endLine]

		pageContent := strings.Join(pageLines, "\n")
		pageContent = strings.TrimPrefix(pageContent, "\n")

		fmt.Print("\033[H\033[2J")

		fmt.Println(pageStyle.Render(pageContent))

		linesPrinted := endLine - startLine
		if linesPrinted < linesPerPage {
			fmt.Print(strings.Repeat("\n", linesPerPage-linesPrinted))
		}

		status := fmt.Sprintf("\nPage %d/%d | Chapter %d/%d", currentPage+1, totalPages, *currentIndex+1, len(spine))
		fmt.Println(docStyle.Render(status))
		//footerText := "[ (n) Next Pg | (p) Prev Pg | (q) Quit ]"
		//fmt.Println(docStyle.Render(footerText))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "n":
			if currentPage < totalPages-1 {
				currentPage++
			} else {
				*currentIndex++
				return
			}
		case "p":
			if currentPage > 0 {
				currentPage--
			} else {
				*currentIndex--
				return
			}
		case "q":
			os.Exit(0)
		default:
			fmt.Println("Invalid input")
		}
	}
}
