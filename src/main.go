package main

import (
	zip2 "archive/zip"
	"encoding/xml"
	"fmt"
	"github.com/taylorskalyo/goreader/epub"
	"io"
	"os"
	"strings"
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

	for _, item := range book.Spine.Itemrefs {
		fileName := findFileName(book, item.ID)
		if fileName == "" {
			continue
		}

		fileStream, err := openZipFile(zip, fileName)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		extractText(fileStream)
		fileStream.Close()
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

func extractText(readableData io.Reader) {
	decoder := xml.NewDecoder(readableData)

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}

		switch token := t.(type) {
		case xml.StartElement:
			if token.Name.Local == "p" {
				fmt.Printf("\n")
			}
		case xml.CharData:
			fmt.Print(string(token))
		}
	}

}
