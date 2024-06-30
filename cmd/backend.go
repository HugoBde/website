package main

import (
	"log"
	"net/http"
	"slices"

	"hugobde.dev/internal/article"
	"hugobde.dev/internal/templates"

	"golang.org/x/exp/maps"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	builder, err := article.NewArticleBuilderWatcher("./website/blog_source")
	if err != nil {
		log.Fatal(err)
	}

	go builder.Run()

	http.HandleFunc("/", homePage)
	http.HandleFunc("/resume", resumePage)
	http.HandleFunc("/blog", blogPageHandler(builder))
	http.HandleFunc("/blog/{id}", articlePageHandler(builder))
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./website/static"))))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./website/index.html")
}

func resumePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./website/resume.html")
}

func blogPageHandler(b *article.ArticleBuilderWatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articleList := maps.Values(b.Articles)
		slices.SortFunc(articleList, func(a *article.BlogArticle, b *article.BlogArticle) int {
			return int(b.PubTime.Sub(a.PubTime))
		})
		templates.BlogTemplate(articleList).Render(r.Context(), w)
	}
}

func articlePageHandler(b *article.ArticleBuilderWatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		article, ok := b.Articles[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("how did you get here buddy?"))
			return
		}

		templates.ArticlePage(article).Render(r.Context(), w)
	}
}
