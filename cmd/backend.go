package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"hugobde.dev/internal/article"
	"hugobde.dev/internal/templates"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
)

// static file server

// blog articles list page

// blog page builder

var BLOG_ARTICLES []*article.BlogArticle = make([]*article.BlogArticle, 0)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file, hope you have environments variables already set :)")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not set")
	}

	serverCert := os.Getenv("SERVER_CERT")
	if serverCert == "" {
		log.Fatal("SERVER_CERT not set")
	}

	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		log.Fatal("SERVER_KEY not set")

	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/", homePage)
	http.HandleFunc("/blog", blogPage)
	http.HandleFunc("/blog/{index}", articlePage)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./website/static"))))

	// go articleBuilder("./blog/source", "./blog")

	articlesResolve("./website/blog_source", "./website/blog")

	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServeTLS(":"+port, serverCert, serverKey, nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./website/index.html")
}

func blogPage(w http.ResponseWriter, r *http.Request) {
	templates.BlogTemplate(BLOG_ARTICLES).Render(r.Context(), w)
}

func articlePage(w http.ResponseWriter, r *http.Request) {
	articleIndex, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	templates.ArticlePage(BLOG_ARTICLES[articleIndex]).Render(r.Context(), w)
}

func articlesResolve(sourceDir string, articleDir string) {
	// Get entries in blog source dir
	sourceDirEntries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatal(err)
	}

	// Get entries in compiled articles dir
	articleDirEntries, err := os.ReadDir(articleDir)
	if err != nil {
		log.Fatal(err)
	}

	// Resolve each blog source with its compiled article
	for _, sourceDirEntry := range sourceDirEntries {
		// Get article name, and ignore entries without .md extension
		articleName, ok := strings.CutSuffix(sourceDirEntry.Name(), ".md")
		if !ok {
			continue
		}

		// Look up matching compiled article
		articleDirEntry, ok := findArticleFile(articleName, articleDirEntries)
		if !ok {
			log.Println("Building article", articleName)
			article, err := article.BuildArticleFromSourceFile(articleName, path.Join(sourceDir, sourceDirEntry.Name()))
			if err != nil {
				log.Fatal(err)
			}
			os.WriteFile(path.Join(articleDir, articleName+".html"), article.HTML, 0644)
			BLOG_ARTICLES = append(BLOG_ARTICLES, article)
			continue
		}

		// Get both files info
		articleFileInfo, err := articleDirEntry.Info()
		sourceFileInfo, err := sourceDirEntry.Info()
		if err != nil {
			log.Fatal(err)
		}

		// If the compiled info is older than the source, rebuild
		if sourceFileInfo.ModTime().Sub(articleFileInfo.ModTime()) > 0 {
			log.Println("Updating article", articleName)
			article, err := article.BuildArticleFromSourceFile(articleName, path.Join(sourceDir, sourceDirEntry.Name()))
			if err != nil {
				log.Fatal(err)
			}
			os.WriteFile(path.Join(articleDir, articleName+".html"), article.HTML, 0644)
			BLOG_ARTICLES = append(BLOG_ARTICLES, article)
			continue
		}

		log.Println("Loading article", articleName)
		article, err := article.ReadArticleFromFile(articleName, path.Join(articleDir, articleFileInfo.Name()), sourceFileInfo.ModTime())
		if err != nil {
			log.Fatal(err)
		}

		BLOG_ARTICLES = append(BLOG_ARTICLES, article)
	}

	// Remove articles for which the source is missing
	for _, articleDirEntry := range articleDirEntries {
		articleName, ok := strings.CutSuffix(articleDirEntry.Name(), ".html")
		if !ok {
			continue
		}

		_, ok = findArticleFile(articleName, sourceDirEntries)
		if ok {
			continue
		}

		os.Remove(path.Join(articleDir, articleDirEntry.Name()))
	}

	// Sort articles in reverse chronological order
	slices.SortFunc[[]*article.BlogArticle](BLOG_ARTICLES, func(a *article.BlogArticle, b *article.BlogArticle) int {
		return int(b.LastModifiedTime.Sub(a.LastModifiedTime))
	})
}

func articleBuilder(sourceDir string, articleDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(sourceDir)
	if err != nil {
		log.Fatal(err)
	}
}

func findArticleFile(articleName string, entries []os.DirEntry) (os.DirEntry, bool) {
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), articleName) {
			return entry, true
		}
	}

	return nil, false
}
