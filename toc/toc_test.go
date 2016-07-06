package toc

import (
	"flag"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	assert "github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// We override TestMain so that we can control the logging level of logrus. If
// we did nothing it would spit out a lot of output even during a normal test
// run. This way it's only verbose if using `go test -v`.
func TestMain(m *testing.M) {
	// we depend on flags so we need to call Parse explicitly (it's otherwise
	// done implicitly inside Run)
	flag.Parse()

	log.SetFormatter(new(log.TextFormatter))

	if testing.Verbose() {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.FatalLevel)
	}

	os.Exit(m.Run())
}

func TestBuildTreeSimple(t *testing.T) {
	node := buildTree([]*header{
		&header{2, "#h0", "Header 0"},
	})
	assert.Equal(t, "ol", node.Data)

	//
	// #h0 (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h0">Header 0</a>`, mustRenderTree(node))
}

func TestBuildTreeComplex(t *testing.T) {
	// The naming in this test isn't particularly good. Headers are named
	// h0-h8, but those conflict with their actual header levels (h2-h5). Just
	// be careful with this one, and you may want to run it with `go test -v`.

	node := buildTree([]*header{
		&header{2, "#h0", "Header 0"},
		&header{2, "#h1", "Header 1"},
		&header{3, "#h2", "Header 2"},
		&header{4, "#h3", "Header 3"},
		&header{5, "#h4", "Header 4"},
		&header{4, "#h5", "Header 5"},
		&header{5, "#h6", "Header 6"},
		&header{3, "#h7", "Header 7"},
		&header{2, "#h8", "Header 8"},
	})

	if testing.Verbose() {
		str, err := renderTree(node)
		assert.NoError(t, err)
		log.Debugf("tree = %+v", str)
	}

	assert.Equal(t, "ol", node.Data)

	//
	// #h0 (<h2>)
	//

	h2Node := node.FirstChild
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h0">Header 0</a>`, mustRenderTree(node))

	//
	// #h1 (<h2>) -- next sibling of #h0
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h1">Header 1</a>`, mustRenderTree(node))

	//
	// #h2 (<h3>) -- child of #h1
	//

	// LastChild because the first is an <a> header link to h1
	h3List := h2Node.LastChild
	assert.NotNil(t, h3List)
	assert.Equal(t, "ol", h3List.Data)

	h3Node := h3List.FirstChild
	assert.NotNil(t, h3Node)
	assert.Equal(t, "li", h3Node.Data)

	node = h3Node.FirstChild
	assert.Equal(t, `<a href="#h2">Header 2</a>`, mustRenderTree(node))

	//
	// #h3 (<h4>) -- child of #h2
	//

	// LastChild because the first is an <a> header link to h2
	h4List := h3Node.LastChild
	assert.NotNil(t, h4List)
	assert.Equal(t, "ol", h4List.Data)

	h4Node := h4List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h4Node.Data)

	node = h4Node.FirstChild
	assert.Equal(t, `<a href="#h3">Header 3</a>`, mustRenderTree(node))

	//
	// #h4 (<h5>) -- child of #h3
	//

	// LastChild because the first is an <a> header link to h3
	h5List := h4Node.LastChild
	assert.NotNil(t, h5List)
	assert.Equal(t, "ol", h5List.Data)

	h5Node := h5List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h5Node.Data)

	node = h5Node.FirstChild
	assert.Equal(t, `<a href="#h4">Header 4</a>`, mustRenderTree(node))

	//
	// #h5 (<h4>) -- next sibiling of #h3
	//

	h4Node = h4Node.NextSibling
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h4Node.Data)

	node = h4Node.FirstChild
	assert.Equal(t, `<a href="#h5">Header 5</a>`, mustRenderTree(node))

	//
	// #h6 (<h5>) -- child of #h5
	//

	// LastChild because the first is an <a> header link to h5
	h5List = h4Node.LastChild
	assert.NotNil(t, h5List)
	assert.Equal(t, "ol", h5List.Data)

	h5Node = h5List.FirstChild
	assert.NotNil(t, h4Node)
	assert.Equal(t, "li", h5Node.Data)

	node = h5Node.FirstChild
	assert.Equal(t, `<a href="#h6">Header 6</a>`, mustRenderTree(node))

	//
	// #h7 (<h2>) -- next sibling of #h2
	//

	h3Node = h3Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h3Node.FirstChild
	assert.Equal(t, `<a href="#h7">Header 7</a>`, mustRenderTree(node))

	//
	// #h8 (<h2>) -- next sibling of #h1
	//

	h2Node = h2Node.NextSibling
	assert.NotNil(t, h2Node)
	assert.Equal(t, "li", h2Node.Data)

	node = h2Node.FirstChild
	assert.Equal(t, `<a href="#h8">Header 8</a>`, mustRenderTree(node))
}

func TestRenderTOC(t *testing.T) {
	content := `
		Intro.

		<h2 id="#h-a">Heading A</h2>

		Content.

		<h3 id="#h-b">Heading B</h3>

		Content

		<h2 id="#h-c">Heading C</h2>

		Content.
	`
	expected := `<ol><li><a href="#h-a">Heading A</a><ol><li><a href="#h-b">Heading B</a></li></ol></li><li><a href="#h-c">Heading C</a></li></ol>`

	rendered, err := RenderTOC(content)
	assert.NoError(t, err)
	assert.Equal(t, expected, rendered)
}

func mustRenderTree(node *html.Node) string {
	str, err := renderTree(node)
	if err != nil {
		panic(err)
	}
	return str
}
