package mtoc

import (
	"testing"

	assert "github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestBuildTree_Basic(t *testing.T) {
	node := buildTree([]*header{
		{2, "#h-a", "Header A"},
	})
	assert.Equal(t, "ol", node.Data)

	//
	// #h-a (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h-a">Header A</a>`, mustRenderTree(node))
}

func TestBuildTree_Complex(t *testing.T) {
	// Be careful with this one, and you may want to run it with `go test -v`.

	node := buildTree([]*header{
		{2, "#h-a", "Header A"},
		{2, "#h-b", "Header B"},
		{3, "#h-c", "Header C"},
		{4, "#h-d", "Header D"},
		{5, "#h-e", "Header E"},
		{4, "#h-f", "Header F"},
		{5, "#h-g", "Header G"},
		{3, "#h-h", "Header H"},
		{2, "#h-i", "Header I"},
	})

	/*
		if testing.Verbose() {
			str, err := renderTree(node)
			assert.NoError(t, err)
			log.Debugf("tree = %+v", str)
		}
	*/

	assert.Equal(t, "ol", node.Data)

	//
	// #h-a (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h-a">Header A</a>`, mustRenderTree(node))

	//
	// #h-b (<h2>) -- next sibling of #h-a
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h-b">Header B</a>`, mustRenderTree(node))

	//
	// #h-c (<h3>) -- child of #h-b
	//

	// LastChild because the first is an <a> header link to h1
	h3List := h2Node.LastChild
	assert.NotNil(t, h3List)
	assert.Equal(t, "ol", h3List.Data)

	h3Node := h3List.FirstChild
	assert.NotNil(t, h3Node)
	assert.Equal(t, "li", h3Node.Data)

	node = h3Node.FirstChild
	assert.Equal(t, `<a href="#h-c">Header C</a>`, mustRenderTree(node))

	//
	// #h-d (<h4>) -- child of #h-c
	//

	// LastChild because the first is an <a> header link to h2
	h4List := h3Node.LastChild
	assert.NotNil(t, h4List)
	assert.Equal(t, "ol", h4List.Data)

	h4Node := h4List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h4Node.Data)

	node = h4Node.FirstChild
	assert.Equal(t, `<a href="#h-d">Header D</a>`, mustRenderTree(node))

	//
	// #h-e (<h5>) -- child of #h-d
	//

	// LastChild because the first is an <a> header link to h3
	h5List := h4Node.LastChild
	assert.NotNil(t, h5List)
	assert.Equal(t, "ol", h5List.Data)

	h5Node := h5List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h5Node.Data)

	node = h5Node.FirstChild
	assert.Equal(t, `<a href="#h-e">Header E</a>`, mustRenderTree(node))

	//
	// #h-f (<h4>) -- next sibiling of #h-d
	//

	h4Node = h4Node.NextSibling
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h4Node.Data)

	node = h4Node.FirstChild
	assert.Equal(t, `<a href="#h-f">Header F</a>`, mustRenderTree(node))

	//
	// #h-g (<h5>) -- child of #h-f
	//

	// LastChild because the first is an <a> header link to h5
	h5List = h4Node.LastChild
	assert.NotNil(t, h5List)
	assert.Equal(t, "ol", h5List.Data)

	h5Node = h5List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h5Node.Data)

	node = h5Node.FirstChild
	assert.Equal(t, `<a href="#h-g">Header G</a>`, mustRenderTree(node))

	//
	// #h-h (<h2>) -- next sibling of #h-c
	//

	h3Node = h3Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h3Node.FirstChild
	assert.Equal(t, `<a href="#h-h">Header H</a>`, mustRenderTree(node))

	//
	// #h-i (<h2>) -- next sibling of #h-b
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h-i">Header I</a>`, mustRenderTree(node))
}

func TestRenderFromHTML_Basic(t *testing.T) {
	content := `
		Intro.

		<h2 id="h-a">Heading A</h2>

		Content.

		<h3 id="h-b">Heading B</h3>

		Content

		<h2 id="h-c"><a href="#h-c">Heading C</a></h2>

		Content.
	`
	expected := `<ol><li><a href="#h-a">Heading A</a><ol><li><a href="#h-b">Heading B</a></li></ol></li><li><a href="#h-c">Heading C</a></li></ol>`

	rendered, err := RenderFromHTML(content)
	assert.NoError(t, err)
	assert.Equal(t, expected, rendered)
}

func TestRenderFromHTML_Empty(t *testing.T) {
	rendered, err := RenderFromHTML("hello")
	assert.NoError(t, err)
	assert.Equal(t, "", rendered)
}

//
// Private
//

func mustRenderTree(node *html.Node) string {
	str, err := renderTree(node)
	if err != nil {
		panic(err)
	}
	return str
}
