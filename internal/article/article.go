package article

import (
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
)

type BlogArticle struct {
	LastModifiedTime time.Time
	HTML             []byte
	Title            string
}

type ArticleBuilderWatcher struct {
	Articles  []*BlogArticle
	watcher   *fsnotify.Watcher
	sourceDir string
	outputDir string
}

func NewArticleBuilderWatcher(sourceDir string, outputDir string) (*ArticleBuilderWatcher, error) {
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
		outputDir: outputDir,
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

	// Get entries in compiled articles dir
	compiledDirEntries, err := os.ReadDir(b.outputDir)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve each blog source with its compiled article
	for _, sourceDirEntry := range sourceDirEntries {
		if !strings.HasSuffix(sourceDirEntry.Name(), ".md") {
			continue
		}

		// Get article name, and ignore entries without .md extension
		title := strings.TrimSuffix(sourceDirEntry.Name(), ".md")

		// Find a matching compiled article
		articleDirEntry := findDirEntry(title, compiledDirEntries)

		// If no match found, build the article
		if articleDirEntry == nil {
			article, err := b.buildArticle(sourceDirEntry.Name())
			if err != nil {
				log.Println(err)
				continue
			}

			os.WriteFile(path.Join(b.outputDir, title+".html"), article.HTML, 0644)
			b.Articles = append(b.Articles, article)
			continue
		}

		// If we find a match, get both files info to compare last modified time
		articleFileInfo, err := articleDirEntry.Info()
		sourceFileInfo, err := sourceDirEntry.Info()
		if err != nil {
			log.Fatal(err)
		}

		// If the compiled file info is older than the source, rebuild
		if sourceFileInfo.ModTime().Sub(articleFileInfo.ModTime()) > 0 {
			article, err := b.buildArticle(sourceDirEntry.Name())
			if err != nil {
				log.Println(err)
				continue
			}
			os.WriteFile(path.Join(b.outputDir, title+".html"), article.HTML, 0644)
			b.Articles = append(b.Articles, article)
			continue
		}

		article, err := b.loadArticle(articleFileInfo.Name(), sourceFileInfo.ModTime())
		if err != nil {
			log.Fatal(err)
		}

		b.Articles = append(b.Articles, article)
		continue
	}

	// Remove articles for which the source is missing
	for _, compiledDirEntry := range compiledDirEntries {
		title, ok := strings.CutSuffix(compiledDirEntry.Name(), ".html")
		if !ok {
			continue
		}

		sourceDirEntry := findDirEntry(title, sourceDirEntries)
		if sourceDirEntry != nil {
			continue
		}

		os.Remove(path.Join(b.outputDir, compiledDirEntry.Name()))
	}

	// Sort articles in reverse chronological order
	slices.SortFunc[[]*BlogArticle](b.Articles, func(a *BlogArticle, b *BlogArticle) int {
		return int(b.LastModifiedTime.Sub(a.LastModifiedTime))
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
			article, err := b.buildArticle(path.Base(event.Name))
			if err != nil {
				log.Println(err)
			}
			if oldArticle := b.findArticle(article.Title); oldArticle != nil {
				b.updateArticle(article)
			} else {
				b.Articles = append(b.Articles, article)
			}

		case fsnotify.Remove:
			b.removeArticle(path.Base(event.Name))

		}

		slices.SortFunc[[]*BlogArticle](b.Articles, func(a *BlogArticle, b *BlogArticle) int {
			return int(b.LastModifiedTime.Sub(a.LastModifiedTime))
		})
	}
}

// Load an already compiled file
func (b *ArticleBuilderWatcher) loadArticle(compiledFile string, lastModifiedTime time.Time) (*BlogArticle, error) {
	title := strings.ReplaceAll(strings.TrimSuffix(compiledFile, ".html"), "_", " ")

	log.Printf("Loading article \"%s\"", title)

	html, err := os.ReadFile(path.Join(b.outputDir, compiledFile))
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

// Build an article from a markdown source file
func (b *ArticleBuilderWatcher) buildArticle(sourceFile string) (*BlogArticle, error) {
	title := strings.ReplaceAll(strings.TrimSuffix(sourceFile, ".md"), "_", " ")

	log.Printf("Building article \"%s\"", title)

	source, err := os.ReadFile(path.Join(b.sourceDir, sourceFile))
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

// Look for article with the same title in the list of existing articles and replace it with the new one
func (b *ArticleBuilderWatcher) updateArticle(newArticle *BlogArticle) bool {
	for i, a := range b.Articles {
		if a.Title == newArticle.Title {
			b.Articles[i] = newArticle
			return true
		}
	}

	return false
}

func (b *ArticleBuilderWatcher) removeArticle(sourceFile string) error {
	title := strings.ReplaceAll(strings.TrimSuffix(sourceFile, ".md"), "_", " ")

	log.Printf("Removing article \"%s\"", title)

	for i, a := range b.Articles {
		if a.Title == title {
			b.Articles = append(b.Articles[:i], b.Articles[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Tried to delete article %s which is not in the list ???", title)
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
