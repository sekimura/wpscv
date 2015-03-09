package common

type TagStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type FetchResult struct {
	Lines   []Line    `json:"lines"`
	Summary []TagStat `json:"summary"`
}

type Line struct {
	Type    string
	Tagname string
	Text    string
	Attr    string
}
