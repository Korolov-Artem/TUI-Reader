package main

import (
	zip2 "archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/taylorskalyo/goreader/epub"
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"unicode/utf8"
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

	for {
		fmt.Print("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n")
		var input string
		var padding string

		item := spine[currentIndex]

		fmt.Println("--------------------------------------------------")
		fmt.Printf("Chapter %d of %d (ID: %s)\n", currentIndex+1, len(spine), item.ID)
		fmt.Println("--------------------------------------------------")

		fileName := findFileName(book, item.ID)
		if fileName == "" {
			fmt.Println("Could not find file for this chapter.")
		} else {
			fileStream, err := openZipFile(zip, fileName)
			if err == nil {
				rawText := extractText(fileStream)
				fileStream.Close()

				if strings.TrimSpace(rawText) == "" {
					fmt.Println("\n[This chapter contains no text (Image or Empty)]")
				} else {
					totalWidth, _, err := term.GetSize(int(os.Stdin.Fd()))
					if err != nil || totalWidth <= 0 {
						totalWidth = 80
					}

					fmt.Println("\nDEBUG RULER (Where the code thinks the screen ends):")
					// Prints a line exactly as wide as the detected width
					fmt.Println("|" + strings.Repeat("-", totalWidth-2) + "|")
					fmt.Printf("Detected Width: %d\n\n", totalWidth)

					textWidth := totalWidth - 10
					if textWidth < 20 {
						textWidth = 20
					}

					leftMarginSize := (totalWidth - textWidth) / 2
					padding = strings.Repeat(" ", leftMarginSize)

					lines := wrapText(rawText, textWidth)

					for _, line := range lines {
						fmt.Println(padding + line)
					}
				}
			}
		}

		fmt.Println("\n--------------------------------------------------")
		fmt.Println("\n" + padding + "[ (n) Next | (p) Prev | (q) Quit ]")

		fmt.Scanln(&input)

		switch input {
		case "n":
			currentIndex++
			if currentIndex >= len(spine) {
				currentIndex = len(spine) - 1
			}
		case "p":
			currentIndex--
			if currentIndex < 0 {
				currentIndex = 0
			}
		case "q":
			return
		default:
			fmt.Println("Invalid input")
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
				token.Name.Local == "br" {
				sb.WriteString("\n")
			}
		case xml.CharData:
			content := string(token)

			content = strings.ReplaceAll(content, "\n", " ")
			content = strings.ReplaceAll(content, "\t", " ")

			sb.WriteString(content)
		}
	}
	return sb.String()
}

func wrapText(text string, lineWidth int) []string {
	var lines []string

	paragraphs := strings.Split(text, "\n")

	for _, paragraph := range paragraphs {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]

		for _, word := range words[1:] {
			currentLen := utf8.RuneCountInString(currentLine)
			wordLen := utf8.RuneCountInString(word)

			if currentLen+1+wordLen <= lineWidth {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}
	return lines
}

func justifyLine(line string, limit int) string {
	words := strings.Fields(line)

	if len(words) == 0 {
		return ""
	}
	if len(words) == 1 {
		return words[0]
	}

	lettersCount := 0

	for _, word := range words {
		lettersCount += utf8.RuneCountInString(word)
	}

	totalSpacesNeeded := limit - lettersCount
	gaps := len(words) - 1
	spacesPerGap := totalSpacesNeeded / gaps
	extraSpaces := totalSpacesNeeded % gaps

	var sb strings.Builder

	for i := 0; i < gaps; i++ {
		sb.WriteString(words[i])

		currentSpaces := spacesPerGap
		if i < extraSpaces {
			currentSpaces++
		}
		sb.WriteString(strings.Repeat(" ", currentSpaces))
	}
	sb.WriteString(words[len(words)-1])
	return sb.String()
}
