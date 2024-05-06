package article

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"gopkg.in/yaml.v3"
)

type BlogArticle struct {
	HTML       []byte
	sourceFile string
	FrontMatter
}

type ArticleBuilderWatcher struct {
	Articles  []*BlogArticle
	watcher   *fsnotify.Watcher
	sourceDir string
}

func NewArticleBuilderWatcher(sourceDir string) (*ArticleBuilderWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(sourceDir)
	if err != nil {
		return nil, err
	}

	return &ArticleBuilderWatcher{
		Articles:  make([]*BlogArticle, 0),
		watcher:   watcher,
		sourceDir: sourceDir,
	}, nil
}

func (b *ArticleBuilderWatcher) Run() {
	b.resume()
	b.watch()
}

func (b *ArticleBuilderWatcher) resume() {

	// Get entries in blog source dir
	sourceDirEntries, err := os.ReadDir(b.sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve each blog source with its compiled article
	for _, sourceDirEntry := range sourceDirEntries {
		if !strings.HasSuffix(sourceDirEntry.Name(), ".md") {
			continue
		}

		article, err := b.buildArticle(path.Join(b.sourceDir, sourceDirEntry.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		b.Articles = append(b.Articles, article)
		continue
	}

	// Sort articles in descending chronological order
	slices.SortFunc(b.Articles, func(a *BlogArticle, b *BlogArticle) int {
		return int(b.PubTime.Sub(a.PubTime))
	})

	log.Printf("Resumption complete: %d article(s)", len(b.Articles))
}

func (b *ArticleBuilderWatcher) watch() {
	for {
		event := <-b.watcher.Events

		// ignore changes to non .md files
		if !strings.HasSuffix(event.Name, ".md") {
			continue
		}

		switch event.Op {
		case fsnotify.Write:
			article, err := b.buildArticle(event.Name)
			if err != nil {
				log.Println(err)
			}
			if oldArticle := b.findArticle(article.Title); oldArticle != nil {
				b.updateArticle(article)
			} else {
				b.Articles = append(b.Articles, article)
			}

		case fsnotify.Remove:
			b.removeArticle(event.Name)

		}

		slices.SortFunc(b.Articles, func(a *BlogArticle, b *BlogArticle) int {
			return int(b.PubTime.Sub(a.PubTime))
		})
	}
}

// Build an article from a markdown source file
func (b *ArticleBuilderWatcher) buildArticle(sourceFile string) (*BlogArticle, error) {
	log.Printf("Building article \"%s\"", sourceFile)

	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, err
	}

	sections := bytes.Split(source, []byte("---\n"))
	if len(sections) != 3 {
		return nil, errors.New("source has invalid frontmatter")
	}

	frontMatter, err := parseFrontMatter(sections[1])
	if err != nil {
		return nil, err
	}

	html := markdown.ToHTML(sections[2], nil, nil)

	article := BlogArticle{
		HTML:        html,
		sourceFile:  sourceFile,
		FrontMatter: frontMatter,
	}

	return &article, nil
}

func parseFrontMatter(source []byte) (FrontMatter, error) {
	var frontMatter FrontMatter
	err := yaml.Unmarshal([]byte(source), &frontMatter)
	if err != nil {
		return FrontMatter{}, err
	}

	return frontMatter, nil
}

// Look for article with the same title in the list of existing articles and replace it with the new one
func (b *ArticleBuilderWatcher) updateArticle(newArticle *BlogArticle) bool {
	for i, a := range b.Articles {
		if a.sourceFile == newArticle.sourceFile {
			b.Articles[i] = newArticle
			return true
		}
	}

	return false
}

func (b *ArticleBuilderWatcher) removeArticle(sourceFile string) error {
	log.Printf("Removing article \"%s\"", sourceFile)

	for i, a := range b.Articles {
		if a.sourceFile == sourceFile {
			b.Articles = append(b.Articles[:i], b.Articles[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Tried to delete article %s which is not in the list ???", sourceFile)
}

func (b *ArticleBuilderWatcher) findArticle(title string) *BlogArticle {
	for _, a := range b.Articles {
		if a.Title == title {
			return a
		}
	}
	return nil
}

func findDirEntry(title string, dirEntries []os.DirEntry) os.DirEntry {
	for _, dirEntry := range dirEntries {
		if strings.HasPrefix(dirEntry.Name(), title) {
			return dirEntry
		}
	}
	return nil
}

type FrontMatter struct {
	Title   string
	ModTime frontMatterTime `yaml:"mod_time"`
	PubTime frontMatterTime `yaml:"pub_time"`
}

type frontMatterTime time.Time

func (t *frontMatterTime) UnmarshalYAML(value *yaml.Node) error {
	parsedTime, err := time.Parse("2006-01-02", value.Value)
	if err != nil {
		return err
	}
	*t = frontMatterTime(parsedTime)
	return nil
}

func (t frontMatterTime) Sub(s frontMatterTime) time.Duration {
	return time.Time(t).Sub(time.Time(s))
}
