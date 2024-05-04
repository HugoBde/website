package article

import (
	"os"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
)

type BlogArticle struct {
	LastModifiedTime time.Time
	HTML             []byte
	Title            string
}

func BuildArticleFromSourceFile(title string, sourceFile string) (*BlogArticle, error) {
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, err
	}

	article := BlogArticle{
		Title:            strings.ReplaceAll(title, "_", " "),
		HTML:             markdown.ToHTML(source, nil, nil),
		LastModifiedTime: time.Now(),
	}

	return &article, nil
}

func ReadArticleFromFile(title string, filename string, lastModifiedTime time.Time) (*BlogArticle, error) {
	html, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	article := BlogArticle{
		Title:            strings.ReplaceAll(title, "_", " "),
		HTML:             html,
		LastModifiedTime: lastModifiedTime,
	}

	return &article, nil
}
