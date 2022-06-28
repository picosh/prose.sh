package internal

import "time"

type MetaData struct {
	PublishAt   *time.Time
	Title       string
	Description string
}

type ParsedText struct {
	Text string
	*MetaData
}

func ParseText(text string) *ParsedText {
	return &ParsedText{
		Text:     text,
		MetaData: &MetaData{},
	}
}
