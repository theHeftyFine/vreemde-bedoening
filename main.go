package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nl/vreemdebedoening/types"
	"os"
	"strings"
)

type Site struct {
	Host string
}

type BlogPost struct {
	Content template.HTML
}

type Index struct {
	Entries []IndexRow
}

type IndexRow struct {
	Name string
	Link string
}

func (s *Site) HandleBlog(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		return
	}
	xmllocation := fmt.Sprintf("./static/%s.xml", params[2])

	err, entry := UnmarshalEntry(xmllocation)

	blogPost := new(BlogPost)
	blogPost.Content = template.HTML(entry.Content.Value)

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

func UnmarshalEntry(xmllocation string) (error, *types.Entry) {
	if _, err := os.Stat(xmllocation); errors.Is(err, os.ErrNotExist) {
		return err, nil
	}
	snippet, err := os.ReadFile(xmllocation)
	if err != nil {
		return err, nil
	}

	var entry types.Entry
	err = xml.Unmarshal(snippet, &entry)
	if err != nil {
		log.Print(err)
		return err, nil
	}
	return nil, &entry
}

func (s *Site) HandleIndex(w http.ResponseWriter, r *http.Request) {
	root := "./static"
	f, err := os.Open(root)
	if err != nil {
		return
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return
	}

	entries := []IndexRow{}
	for _, file := range files {
		err, entry := UnmarshalEntry(root + "/" + file.Name())
		if err != nil {
			log.Print(err)
			return
		}
		indexRow := new(IndexRow)
		indexRow.Name = entry.Title
		indexRow.Link = strings.Replace(entry.Link.Href, "${HOST}", s.Host, -1)
		entries = append(entries, *indexRow)
	}
	index := new(Index)
	index.Entries = entries
	index.Entries = entries
	tmpl, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		return
	}

	err = tmpl.Execute(w, index)
	if err != nil {
		return
	}
}

func (s *Site) HandleFeed(w http.ResponseWriter, r *http.Request) {
	feedTemplate, err := os.ReadFile("./templates/feed.xml")
	if err != nil {
		log.Print(err)
		return
	}

	var feed types.Feed
	err = xml.Unmarshal(feedTemplate, &feed)
	if err != nil {
		log.Print(err)
		return
	}

	root := "./static"
	f, err := os.Open(root)
	if err != nil {
		return
	}
	files, err := f.Readdir(-1)
	if err != nil {
		return
	}

	entries := []types.Entry{}
	for _, file := range files {
		err, entry := UnmarshalEntry(root + "/" + file.Name())
		if err != nil {
			log.Print(err)
			return
		}
		entries = append(entries, *entry)
	}

	feed.Entries = entries

	feed.SetHost(s.Host)

	out, err := xml.MarshalIndent(feed, " ", "  ")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"))

	w.Write(out)
}

func (s *Site) serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	})
}

func (s *Site) StaticFilesHandler(path http.Dir) http.Handler {
	handler := http.FileServer(path)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		handler.ServeHTTP(w, r)
	})
}

func (s *Site) IndexMiddleHandler(path http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.HandleIndex(w, r)
			return
		}
		http.FileServer(path).ServeHTTP(w, r)
	})
}

func main() {

	host, set := os.LookupEnv("HTTP_HOST")
	if !set {
		log.Fatal("HTTP_HOST is not set")
	}

	site := new(Site)
	site.Host = host
	site.Host = host
	indexHandler := site.IndexMiddleHandler("./public/")

	http.Handle("/", indexHandler)
	http.HandleFunc("/blog/", site.HandleBlog)
	http.HandleFunc("/feed", site.HandleFeed)

	// Start the server on port 8080
	log.Println("Starting server on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
