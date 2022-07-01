package internal

import (
	"bytes"
	"fmt"
	"time"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type MetaData struct {
	PublishAt   *time.Time
	Title       string
	Description string
	Nav         []Link
}

type ParsedText struct {
	Html string
	*MetaData
}

func toString(obj interface{}) string {
	if obj == nil {
		return ""
	}
	return obj.(string)
}

func toLinks(obj interface{}) []Link {
	links := []Link{}
	if obj == nil {
		return links
	}

	raw := obj.(map[interface{}]interface{})
	for k, v := range raw {
		links = append(links, Link{
			Text: k.(string),
			URL:  v.(string),
		})
	}

	return links
}

func ParseText(text string) *ParsedText {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, meta.Meta),
	)
	context := parser.NewContext()
	if err := md.Convert([]byte(text), &buf, parser.WithContext(context)); err != nil {
		fmt.Println(err)
		return &ParsedText{}
	}
	metaData := meta.Get(context)

	publishAt := time.Now()
	var err error
	date := toString(metaData["date"])
	if date != "" {
		publishAt, err = time.Parse("2006-01-02", date)
		if err != nil {
			fmt.Println(err)
		}
	}

	nav := toLinks(metaData["nav"])

	return &ParsedText{
		Html: buf.String(),
		MetaData: &MetaData{
			PublishAt:   &publishAt,
			Title:       toString(metaData["title"]),
			Description: toString(metaData["description"]),
			Nav:         nav,
		},
	}
}
