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

	wpscv "github.com/sekimura/wpscv/common"
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

// flatten HTML tokens to lines with tag usage data
func flatten(r io.Reader) (lines []wpscv.Line, stats []wpscv.TagStat) {
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
			l := wpscv.Line{
				Type: "Text",
				Text: "<!DOCTYPE " + t.Data + ">",
			}
			lines = append(lines, l)
		case html.CommentToken:
			l := wpscv.Line{
				Type: "Text",
				Text: "<!--" + t.Data + "-->",
			}
			lines = append(lines, l)
		case html.SelfClosingTagToken, html.StartTagToken:
			m[t.Data]++
			l := wpscv.Line{
				Type:    "Tag",
				Attr:    attrString(t.Attr),
				Tagname: t.Data,
			}
			lines = append(lines, l)
		case html.EndTagToken:
			l := wpscv.Line{
				Type:    "EndTag",
				Tagname: t.Data,
			}
			lines = append(lines, l)
		case html.TextToken:
			if len(strings.Fields(t.Data)) != 0 {
				l := wpscv.Line{
					Type: "Text",
					Text: t.Data,
				}
				lines = append(lines, l)
			}
		}
	}

	for k, v := range m {
		stats = append(stats, wpscv.TagStat{
			Name:  k,
			Count: v,
		})
	}
	return
}

func fetch(u string) (*wpscv.FetchResult, error) {
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	types := strings.Split(res.Header.Get("Content-Type"), ";")
	if strings.ToLower(types[0]) != "text/html" {
		log.Println(res.Header.Get("Content-Type"))
		return nil, fmt.Errorf("Invalid Content-Type")
	}
	fr := &wpscv.FetchResult{}
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
	w.Header().Set("Content-Type", "application/json")
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
