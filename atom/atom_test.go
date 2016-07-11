package atom

import (
	"bytes"
	"encoding/xml"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestFeed(t *testing.T) {
	f := &Feed{
		Title: "My Blog",
		ID:    "blog-unique-id",

		Links: []*Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://example.com"},
			{Rel: "alternate", Type: "application/atom+xml", Href: "https://example.com"},
		},

		Entries: []*Entry{
			{
				Title:   "My Entry",
				Content: &EntryContent{"Entry content."},
				Link:    &Link{Href: "https://example.com"},
				ID:      "tag:site:article-id",

				AuthorName: "Jane Doe",
				AuthorURI:  "https://example.com",
			},
		},
	}

	var b bytes.Buffer
	err := f.Encode(&b, "")
	assert.NoError(t, err)

	assert.Equal(t,
		`<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
			`<feed xml:lang="en-US" xmlns="http://www.w3.org/2005/Atom">`+
			`<title>My Blog</title>`+
			`<id>blog-unique-id</id>`+
			`<updated>0001-01-01T00:00:00Z</updated>`+
			`<link rel="self" type="application/atom+xml" href="https://example.com"></link>`+
			`<link rel="alternate" type="application/atom+xml" href="https://example.com"></link>`+
			`<entry>`+
			`<title>My Entry</title>`+
			`<content><![CDATA[Entry content.]]></content>`+
			`<published>0001-01-01T00:00:00Z</published>`+
			`<updated>0001-01-01T00:00:00Z</updated>`+
			`<link href="https://example.com"></link>`+
			`<id>tag:site:article-id</id>`+
			`<author>`+
			`<name>Jane Doe</name>`+
			`<uri>https://example.com</uri>`+
			`</author>`+
			`</entry>`+
			`</feed>`,
		b.String())
}

func TestLink(t *testing.T) {
	link := &Link{Rel: "self", Type: "application/atom+xml", Href: "https://example.com"}

	var b bytes.Buffer
	enc := xml.NewEncoder(&b)
	err := enc.Encode(link)
	assert.NoError(t, err)

	assert.Equal(t,
		`<link rel="self" type="application/atom+xml" href="https://example.com"></link>`,
		b.String())
}
