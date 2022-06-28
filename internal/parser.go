package internal

import (
	"bytes"
	"fmt"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type MetaData struct {
	PublishAt   *time.Time
	Title       string
	Description string
}

type ParsedText struct {
	Html string
	*MetaData
}

func ParseText(text string) *ParsedText {
	var buf bytes.Buffer
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
	)
	if err := md.Convert([]byte(text), &buf); err != nil {
		fmt.Println(err)
		return &ParsedText{}
	}

	return &ParsedText{
		Html:     buf.String(),
		MetaData: &MetaData{},
	}
}
