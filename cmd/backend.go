package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"hugobde.dev/internal/article"
	"hugobde.dev/internal/templates"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	builder, err := article.NewArticleBuilderWatcher("./website/blog_source")
	if err != nil {
		log.Fatal(err)
	}

	go builder.Run()

	http.HandleFunc("/", homePage)
	http.HandleFunc("/blog", blogPageHandler(builder))
	http.HandleFunc("/blog/{index}", articlePageHandler(builder))
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./website/static"))))

	log.Println("Listening on :" + port)
	log.Fatal(http.ListenAndServeTLS(":"+port, serverCert, serverKey, nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./website/index.html")
}

func blogPageHandler(b *article.ArticleBuilderWatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates.BlogTemplate(b.Articles).Render(r.Context(), w)
	}
}

func articlePageHandler(b *article.ArticleBuilderWatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articleIndex, err := strconv.Atoi(r.PathValue("index"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}

		if articleIndex >= len(b.Articles) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("how did you get here buddy?"))
			return
		}

		templates.ArticlePage(b.Articles[articleIndex]).Render(r.Context(), w)
	}
}
