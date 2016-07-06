package toc

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestBuildTreeSimple(t *testing.T) {
	h0 := &header{2, "#h0", "Header 0"}

	node := buildTree([]*header{h0})
	assert.Equal(t, "ol", node.Data)

	//
	// h0 (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "a", node.Data)
	assert.Equal(t, []html.Attribute{html.Attribute{"", "href", "#h0"}}, node.Attr)

	node = node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "Header 0", node.Data)
}

func TestBuildTreeComplex(t *testing.T) {
	h0 := &header{2, "#h0", "Header 0"}
	h1 := &header{2, "#h1", "Header 1"}
	h2 := &header{3, "#h2", "Header 2"}
	h3 := &header{4, "#h3", "Header 3"}
	h4 := &header{2, "#h4", "Header 4"}

	node := buildTree([]*header{h0, h1, h2, h3, h4})
	assert.Equal(t, "ol", node.Data)

	//
	// h0 (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "a", node.Data)
	assert.Equal(t, []html.Attribute{html.Attribute{"", "href", "#h0"}}, node.Attr)

	node = node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "Header 0", node.Data)

	//
	// h1 (<h2>) -- next sibling of h0
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "a", node.Data)
	assert.Equal(t, []html.Attribute{html.Attribute{"", "href", "#h1"}}, node.Attr)

	node = node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "Header 1", node.Data)

	//
	// h4 (<h2>) -- next sibling of h1
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "a", node.Data)
	assert.Equal(t, []html.Attribute{html.Attribute{"", "href", "#h4"}}, node.Attr)

	node = node.FirstChild
	assert.NotNil(t, node)
	assert.Equal(t, "Header 4", node.Data)
}
