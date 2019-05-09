package mtoc

import (
	"bytes"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

type header struct {
	level int
	id    string
	title string
}

// RenderFromHTML extracts a structure from the given HTML content and renders
// a corresponding table of contents as an HTML string.
func RenderFromHTML(content string) (string, error) {
	var headers []*header

	matches := headerRegexp.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		level, err := strconv.Atoi(match[1])
		if err != nil {
			return "", errors.Wrap(err, "Error extracting header level")
		}

		headers = append(headers, &header{level, "#" + match[2], match[4]})
	}

	node := buildTree(headers)

	// Handle an article that doesn't have any TOC.
	if node == nil {
		return "", nil
	}

	return renderTree(node)
}

//
// Private
//

var headerRegexp = regexp.MustCompile(`<h([0-9]) id="(.*?)">(<a.*?>)?(.*?)(</a>)?</h[0-9]>`)

func buildTree(headers []*header) *html.Node {
	if len(headers) < 1 {
		return nil
	}

	listNode := &html.Node{Data: "ol", Type: html.ElementNode}

	// keep a reference back to the top of the list
	topNode := listNode

	listItemNode := &html.Node{Data: "li", Type: html.ElementNode}
	listNode.AppendChild(listItemNode)

	// This basically helps us track whether we've insert multiple headers on
	// the same level in a row. If we did, we need to create a new list item
	// for each.
	needNewListNode := false

	var level int
	if len(headers) > 0 {
		level = headers[0].level
	}

	for _, header := range headers {
		if header.level > level {
			// indent

			// for each level indented, create a new nested list
			for i := 0; i < (header.level - level); i++ {
				listNode = &html.Node{Data: "ol", Type: html.ElementNode}
				listItemNode.AppendChild(listNode)
			}

			needNewListNode = true

			level = header.level
		} else if header.level < level {
			// dedent

			// for each level outdented, move up two parents, one for list item
			// and one for list
			for i := 0; i < (level - header.level); i++ {
				listItemNode = listNode.Parent
				listNode = listItemNode.Parent
			}

			level = header.level
		}

		if needNewListNode {
			listItemNode = &html.Node{Data: "li", Type: html.ElementNode}
			listNode.AppendChild(listItemNode)
			needNewListNode = false
		}

		contentNode := &html.Node{Data: header.title, Type: html.TextNode}

		linkNode := &html.Node{
			Data: "a",
			Attr: []html.Attribute{
				{Namespace: "", Key: "href", Val: header.id},
			},
			Type: html.ElementNode,
		}
		linkNode.AppendChild(contentNode)
		listItemNode.AppendChild(linkNode)

		needNewListNode = true
	}

	return topNode
}

func renderTree(node *html.Node) (string, error) {
	var b bytes.Buffer
	err := html.Render(&b, node)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
