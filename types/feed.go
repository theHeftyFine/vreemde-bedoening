package types

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Entry struct {
	XMLName xml.Name `xml:"entry"`
	Link    Link     `xml:"link"`
	Title   string   `xml:"title"`
	Id      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Content Content  `xml:"content"`
}

type Link struct {
	XMLName xml.Name `xml:"link"`
	Href    string   `xml:"href,attr"`
}

type Content struct {
	XMLName xml.Name `xml:"content"`
	Type    string   `xml:"type,attr"`
	Value   string   `xml:",innerxml"`
}

type Author struct {
	XMLName xml.Name `xml:"author"`
	Name    string   `xml:"name"`
}

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Xmlns   string   `xml:"xmlns,attr"`
	Title   string   `xml:"title"`
	Link    Link     `xml:"link"`
	Id      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Author  Author   `xml:"author"`
	Entries []Entry  `xml:"entry"`
}

func (f *Feed) String() string {
	return fmt.Sprintf("title=%s, xmlns=%s, id=%s, link=%s", f.Title, f.Xmlns, f.Id, f.Link.Href)
}

func (f *Feed) SetHost(host string) {
	f.Link.Href = strings.Replace(f.Link.Href, "${HOST}", host, -1)
	f.Id = strings.Replace(f.Id, "${HOST}", host, -1)
	for i := range f.Entries {
		f.Entries[i].SetHost(host)
	}
}

func (e *Entry) String() string {
	return fmt.Sprintf("id=%s, title=%s, link=%s, Type=%s, Content=%s", e.Id, e.Title, e.Link.Href, e.Content.Type, e.Content.Value)
}

func (e *Entry) SetHost(host string) {
	e.Link.Href = strings.Replace(e.Link.Href, "${HOST}", host, -1)
	e.Id = strings.Replace(e.Id, "${HOST}", host, -1)
}
