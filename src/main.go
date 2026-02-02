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
	fmt.Println("--------------------------------------------------")

	spine := book.Spine.Itemrefs
	currentIndex := 0

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("\033[H\033[2J")

		totalWidth, totalHeight, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil || totalWidth <= 0 {
			totalWidth = 80
			totalHeight = 24
		}

		verticalMargin := 5
		linesPerPage := totalHeight - verticalMargin
		if linesPerPage < 1 {
			linesPerPage = 1
		}

		item := spine[currentIndex]

		docStyle := lipgloss.NewStyle().
			Width(totalWidth).
			Align(lipgloss.Center)

		textStyle := lipgloss.NewStyle().
			Width(min(80, totalWidth-10)).
			Align(lipgloss.Left)

		fmt.Printf("Chapter %d of %d (ID: %s)\n", currentIndex+1, len(spine), item.ID)

		fileName := findFileName(book, item.ID)
		if fileName == "" {
			fmt.Println("Could not find file for this chapter.")
			fmt.Println("\nPress 'n' for next, 'p' for prev.")
			handleNavigation(reader, &currentIndex, len(spine))
			continue
		} else {
			fileStream, err := openZipFile(zip, fileName)
			if err != nil {
				handleNavigation(reader, &currentIndex, len(spine))
				continue
			}
			rawText := extractText(fileStream)
			formattedText := strings.ReplaceAll(rawText, "\t", "    ")
			fileStream.Close()

			if strings.TrimSpace(rawText) == "" {
				fmt.Println("\n[This chapter contains no text (Image or Empty)]")
				fmt.Println(docStyle.Render("\n[ Press (n) Next or (p) Prev ]"))
				handleNavigation(reader, &currentIndex, len(spine))
				continue
			} else {
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
					fmt.Println(docStyle.Render(pageContent))

					linesPrinted := endLine - startLine
					if linesPrinted < linesPerPage {
						fmt.Print(strings.Repeat("\n", linesPerPage-linesPrinted))
					}

					status := fmt.Sprintf("\nPage %d/%d | Chapter %d/%d", currentPage+1, totalPages, currentIndex+1, len(spine))
					fmt.Println(docStyle.Render(status))

					footerText := "[ (n) Next Pg | (p) Prev Pg | (N) Next Chap | (P) Prev Chap | (q) Quit ]"
					fmt.Println(docStyle.Render(footerText))

					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)

					shouldBreak := false

					switch input {
					case "n":
						if currentPage < totalPages-1 {
							currentPage++
						} else {
							currentIndex++
							shouldBreak = true
						}
					case "p":
						if currentPage > 0 {
							currentPage--
						} else {
							currentIndex--
							shouldBreak = true
						}
					case "N":
						currentIndex++
						shouldBreak = true
					case "P":
						currentIndex--
						shouldBreak = true
					case "q":
						return
					default:
						fmt.Println("Invalid input")
					}
					if shouldBreak {
						break
					}
				}
			}
			if currentIndex >= len(spine) {
				currentIndex = len(spine) - 1
			}
			if currentIndex < 0 {
				currentIndex = 0
			}
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
	case "n", "N":
		*currentIndex++
	case "p", "P":
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
