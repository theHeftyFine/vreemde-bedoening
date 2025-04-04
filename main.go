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
	parts := strings.Split(r.URL.Path, "/")
	params := []string{}
	for _, s := range parts {
		if len(s) > 0 {
			params = append(params, s)
		}
	}
	if len(params) != 2 {
		s.HandleIndex(w, r)
		return
	}
	xmllocation := fmt.Sprintf("./static/%s.xml", params[1])

	err, entry := UnmarshalEntry(xmllocation)

	blogPost := new(BlogPost)
	blogPost.Content = template.HTML(entry.Content.Value)

	tmpl, err := template.ParseFiles("./templates/layout/base.html", "./templates/blog_page.html")
	if err != nil {
		log.Print("Couldn't parse template")
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", blogPost)
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
		fmt.Println(err)
		return
	}
	files, err := f.Readdir(-1)
	if err != nil {
		fmt.Println(err)
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
	tmpl, err := template.ParseFiles("./templates/layout/base.html", "./templates/index.html")
	if err != nil {
		log.Print(err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", index)
	if err != nil {
		log.Print(err)
		return
	}
}

func (s *Site) HandleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/layout/base.html", "./templates/home.html")
	if err != nil {
		fmt.Println("Error parsing template: %s", err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		fmt.Println("Error rendering template: %s", err)
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

func main() {

	host, set := os.LookupEnv("HTTP_HOST")
	if !set {
		log.Fatal("HTTP_HOST is not set")
	}

	site := new(Site)
	site.Host = host
	site.Host = host
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("./public/"))

	mux.HandleFunc("GET /", site.HandleHome)
	mux.HandleFunc("GET /blog/", site.HandleBlog)
	mux.HandleFunc("GET /feed", site.HandleFeed)
	mux.Handle("GET /style.css", fs)
	mux.Handle("GET /vikingS.png", fs)

	// Start the server on port 8080
	log.Println("Starting server on http://localhost:8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
