package internal

import (
	"bytes"
	"fmt"
	"time"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
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

func toLinks(obj interface{}) ([]Link, error) {
	links := []Link{}
	if obj == nil {
		return links, nil
	}

	switch raw := obj.(type) {
	case map[interface{}]interface{}:
		for k, v := range raw {
			links = append(links, Link{
				Text: k.(string),
				URL:  v.(string),
			})
		}
	default:
		return links, fmt.Errorf("unsupported type for `nav` variable: %T", raw)
	}

	return links, nil
}

func ParseText(text string) (*ParsedText, error) {
	var buf bytes.Buffer
	hili := highlighting.NewHighlighting(
		highlighting.WithStyle("dracula"),
		highlighting.WithFormatOptions(
			html.WithLineNumbers(true),
		),
	)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
			hili,
		),
	)
	context := parser.NewContext()
	if err := md.Convert([]byte(text), &buf, parser.WithContext(context)); err != nil {
		return &ParsedText{}, err
	}
	metaData := meta.Get(context)

	publishAt := time.Now()
	var err error
	date := toString(metaData["date"])
	if date != "" {
		publishAt, err = time.Parse("2006-01-02", date)
		if err != nil {
			return &ParsedText{}, err
		}
	}

	nav, err := toLinks(metaData["nav"])
	if err != nil {
		return &ParsedText{}, err
	}

	return &ParsedText{
		Html: buf.String(),
		MetaData: &MetaData{
			PublishAt:   &publishAt,
			Title:       toString(metaData["title"]),
			Description: toString(metaData["description"]),
			Nav:         nav,
		},
	}, nil
}
