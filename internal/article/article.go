package article

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
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
	Articles  map[string]*BlogArticle
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
		Articles:  make(map[string]*BlogArticle, 0),
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

		b.Articles[article.ID] = article
		continue
	}

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
			} else {
				b.Articles[article.ID] = article
			}

		case fsnotify.Remove:
			b.removeArticle(event.Name)

		}

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
	if frontMatter.Title == "" {
		return FrontMatter{}, errors.New("Missing Title in front matter")
	}
	if frontMatter.ID == "" {
		return FrontMatter{}, errors.New("Missing ID in front matter")
	}

	return frontMatter, nil
}

func (b *ArticleBuilderWatcher) removeArticle(sourceFile string) error {
	log.Printf("Removing article \"%s\"", sourceFile)

	for k, a := range b.Articles {
		if a.sourceFile == sourceFile {
			delete(b.Articles, k)
			return nil
		}
	}
	return fmt.Errorf("Tried to delete article %s which is not in the list ???", sourceFile)
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
	ID      string          `yaml:"id"`
	Title   string          `yaml:"title"`
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
