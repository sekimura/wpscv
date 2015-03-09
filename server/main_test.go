package main

import (
	"reflect"
	"strings"
	"testing"

	wpscv "github.com/sekimura/wpscv/common"
	"golang.org/x/net/html"
)

type attrStringTest struct {
	desc     string
	input    []html.Attribute
	expected string
}

var attrStringTests = []attrStringTest{
	{
		"empty",
		[]html.Attribute{},
		"",
	},
	{
		"sane",
		[]html.Attribute{
			{
				Namespace: "",
				Key:       "foo",
				Val:       "bar",
			},
		},
		` foo="bar"`,
	},
	{
		"double",
		[]html.Attribute{
			{
				Namespace: "",
				Key:       "foo",
				Val:       "a",
			},
			{
				Namespace: "",
				Key:       "bar",
				Val:       "b",
			},
		},
		` foo="a" bar="b"`,
	},
}

func TestAttString(t *testing.T) {
	for _, test := range attrStringTests {
		actual := attrString(test.input)
		if actual != test.expected {
			t.Error("did not match expected value:", actual, ",", test.expected)
		}
	}
}

type flattenTestExpected struct {
	lines []wpscv.Line
	stats []wpscv.TagStat
}
type flattenTest struct {
	desc     string
	input    string
	expected flattenTestExpected
}

var flattenTests = []flattenTest{
	{
		"simple",
		`<!DOCTYPE html><html lang="en"><body></body></html>`,
		flattenTestExpected{
			[]wpscv.Line{
				{
					Type:    "Text",
					Tagname: "",
					Text:    "<!DOCTYPE html>",
					Attr:    "",
				}, {
					Type:    "Tag",
					Tagname: "html",
					Text:    "",
					Attr:    ` lang="en"`,
				}, {
					Type:    "Tag",
					Tagname: "body",
					Text:    "",
					Attr:    "",
				}, {
					Type:    "EndTag",
					Tagname: "body",
					Text:    "",
					Attr:    "",
				}, {
					Type:    "EndTag",
					Tagname: "html",
					Text:    "",
					Attr:    "",
				},
			},
			[]wpscv.TagStat{
				{
					Name:  "html",
					Count: 1,
				},
				{
					Name:  "body",
					Count: 1,
				},
			},
		},
	},
}

func TestFlatten(t *testing.T) {
	for _, test := range flattenTests {
		r := strings.NewReader(test.input)
		lines, stats := flatten(r)
		if !reflect.DeepEqual(lines, test.expected.lines) {
			t.Error("did not match expected value:", lines, ",", test.expected.lines)
		}
		if !reflect.DeepEqual(stats, test.expected.stats) {
			t.Error("did not match expected value:", stats, ",", test.expected.stats)
		}
	}
}
