package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func attrString(attr []html.Attribute) string {
	buf := new(bytes.Buffer)
	for _, a := range attr {
		buf.WriteByte(' ')
		buf.WriteString(a.Key)
		buf.WriteString(`="`)
		buf.WriteString(html.EscapeString(a.Val))
		buf.WriteByte('"')
	}
	return buf.String()
}

func flatten(r io.Reader) (lines []Line, stats []TagStat) {
	m := make(map[string]int)
	z := html.NewTokenizer(r)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if err := z.Err(); err != nil && err != io.EOF {
				return
			}
			break
		}

		t := z.Token()
		switch tt {
		case html.DoctypeToken:
			l := Line{
				Type: "Doctype",
				Text: "<!DOCTYPE " + t.Data + ">",
			}
			lines = append(lines, l)
		case html.CommentToken:
			l := Line{
				Type: "Comment",
				Text: "<!--" + t.Data + "-->",
			}
			lines = append(lines, l)
		case html.SelfClosingTagToken, html.StartTagToken:
			m[t.Data]++
			l := Line{
				Type:    "Tag",
				Attr:    attrString(t.Attr),
				Tagname: t.Data,
			}
			lines = append(lines, l)
		case html.EndTagToken:
			l := Line{
				Type:    "EndTag",
				Tagname: t.Data,
			}
			lines = append(lines, l)
		case html.TextToken:
			if len(strings.Fields(t.Data)) != 0 {
				l := Line{
					Type: "Text",
					Text: t.Data,
				}
				lines = append(lines, l)
			}
		}
	}

	for k, v := range m {
		stats = append(stats, TagStat{
			Name:  k,
			Count: v,
		})
	}
	return
}

func fetch(u string) (*FetchResult, error) {
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	fr := &FetchResult{}
	lines, stats := flatten(res.Body)
	fr.Lines = lines
	fr.Summary = stats
	return fr, nil
}

func fetcherHandler(w http.ResponseWriter, r *http.Request) {
	u := r.FormValue("u")
	log.Println("Fetching: ", u)
	fr, err := fetch(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		log.Println("Fetching Error: ", err.Error())
		return
	}
	b, err := json.Marshal(fr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err.Error())
		log.Println("Fetching Error: ", err.Error())
		return
	}
	log.Println("Fetching Success: ", u)
	w.Write(b)
}

func main() {
	addr := flag.String("addr", ":8999", "listen address")
	staticDir := flag.String("dir", "", "directory for static files")
	flag.Parse()

	if *staticDir == "" {
		d, err := os.Getwd()
		if err != nil {
			log.Fatal("could not get the current directory path")
		}
		d += "/static"
		staticDir = &d
	}
	log.Println("Listen Address: ", *addr)
	log.Println("Static Dir: ", *staticDir)

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))
	http.HandleFunc("/api/fetch", fetcherHandler)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
