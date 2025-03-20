package main

import (
	"net/http"
	"log"
	"html/template"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	"errors"
)

type BlogPost struct {
	Content template.HTML
}

type Index struct {
	Entries []string
}

func HandleBlog(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		return
	}	

	location := fmt.Sprintf("./public/blog/%s.html", params[2])
	log.Print(location)
	if _, err := os.Stat(location); errors.Is(err, os.ErrNotExist) {
		log.Print("Non existing blogpost")
		return
	}

	content, err := ioutil.ReadFile(location)
	if err != nil {
		return
	}

	blogPost := new(BlogPost)
	blogPost.Content = template.HTML(string(content))

	tmpl, err := template.ParseFiles("./templates/blog_page.html")
	if err != nil {
		log.Print("Couldn't parse template")
		return
	}

	err = tmpl.Execute(w, blogPost)
	if err != nil {
		log.Print("Could't render template")
		return
	}

}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	root := "./public/blog"
	f, err := os.Open(root)
	if err != nil {
		return
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return
	}

	entries := []string{}
	for _, file := range files {
		parts := strings.Split(file.Name(), ".")
		if len(parts) != 2 {
			return
		}
		log.Print(parts[0])
		entries = append(entries, parts[0])
	}
	index := new(Index)
	index.Entries = entries;

	tmpl, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		return
	}

	err = tmpl.Execute(w, index)
	if err != nil {
		return
	}
}

func serveSingle(pattern string, filename string) {
    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, filename)
    })
}

func StaticFilesHandler(path http.Dir) http.Handler {
	handler := http.FileServer(path)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        handler.ServeHTTP(w, r)
    })
}

func IndexMiddleHandler(path http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			HandleIndex(w, r)
			return
		}
		http.FileServer(path).ServeHTTP(w, r)
	})
}

func main() {
	indexHandler := IndexMiddleHandler("./public/")
	
	http.Handle("/", indexHandler)
	http.HandleFunc("/blog/", HandleBlog)

	// Start the server on port 8080
	log.Println("Starting server on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}