package toc

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"golang.org/x/net/html"
)

type header struct {
	level int
	id    string
	title string
}

var headerRegexp *regexp.Regexp

func init() {
	headerRegexp = regexp.MustCompile(`<h([0-9]) id="(.*)">(.*)</h[0-9]>`)
}

func RenderTOC(content string) (string, error) {
	var headers []*header

	matches := headerRegexp.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		level, err := strconv.Atoi(match[0])
		if err != nil {
			return "", fmt.Errorf("Couldn't extract header level: %v", err.Error())
		}

		headers = append(headers, &header{level, match[1], match[2]})
	}

	parent := buildTree(headers)

	var b bytes.Buffer
	err := html.Render(&b, parent)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func buildTree(headers []*header) *html.Node {
	if len(headers) < 1 {
		return nil
	}

	parent := &html.Node{Data: "ol"}
	node := parent

	var level int
	if len(headers) > 0 {
		level = headers[0].level
	}

	for _, header := range headers {
		if header.level > level {
			// indent

			listItemNode := &html.Node{Data: "li"}
			node.AppendChild(listItemNode)

			listNode := &html.Node{Data: "ol"}
			listItemNode.AppendChild(listNode)

			node = listNode

		} else if header.level < level {
			// outdent

			// move up two parents, one for list item and one for list
			node = node.Parent
			node = node.Parent
		}

		contentNode := &html.Node{Data: header.title}

		linkNode := &html.Node{
			Data: "a",
			Attr: []html.Attribute{
				html.Attribute{"", "href", header.id},
			},
		}
		linkNode.AppendChild(contentNode)

		listItemNode := &html.Node{Data: "li"}
		listItemNode.AppendChild(linkNode)

		// Attach new content into the current list.
		node.AppendChild(listItemNode)
	}

	return parent
}
