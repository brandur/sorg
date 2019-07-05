package smarkdown

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCollapseHTML(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>", collapseHTML(`
<p>
  <strong>strong</strong>
</p>`))
}

func TestRender(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", Render("**strong**", nil))
}

func TestTransformCodeWithLanguagePrefix(t *testing.T) {
	assert.Equal(t,
		`<code class="language-ruby">`,
		transformCodeWithLanguagePrefix(`<code class="ruby">`, nil),
	)
}

func TestTransformSections(t *testing.T) {
	assert.Equal(t, `<section class="section-class">

hello

</section>
`,
		transformSections(`!section class="section-class"

hello

!/section
`, nil))

	// Test once through the full render function as well so that we can make
	// sure that it still works even after content has been garbled by
	// Markdown.
	assert.Equal(t, `<section class="section-class">

<p>hello</p>

</section>
`,
		Render(`!section class="section-class"

hello

!/section
`, nil))
}

func TestTransformFigures(t *testing.T) {
	assert.Equal(t, `
<figure>
  <p><a href="fig-src"><img src="fig-src" class="overflowing"></a></p>
  <figcaption>fig-caption</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src" caption="fig-caption"`, nil),
	)

	// .png links to "@2x" version of the source
	assert.Equal(t, `
<figure>
  <p><a href="fig-src@2x.png"><img src="fig-src.png" class="overflowing"></a></p>
  <figcaption>fig-caption</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src.png" caption="fig-caption"`, nil),
	)

	// .svg doesn't link to "@2x"
	assert.Equal(t, `
<figure>
  <p><a href="fig-src.svg"><img src="fig-src.svg" class="overflowing"></a></p>
  <figcaption>fig-caption</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src.svg" caption="fig-caption"`, nil),
	)

	assert.Equal(t, `
<figure>
  <p><a href="fig-src"><img src="fig-src" class="overflowing"></a></p>
  <figcaption>Caption with some "" quote.</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src" caption="Caption with some \"\" quote."`, nil),
	)
}

func TestTransformFootnotes(t *testing.T) {
	assert.Equal(t, `
<p>This is a reference <sup id="footnote-1-source"><a href="#footnote-1">1</a></sup> to a footnote <sup id="footnote-2-source"><a href="#footnote-2">2</a></sup>.</p>

<p>Not footnote: KEYS[1].</p>


<div id="footnotes">
  <p><sup id="footnote-1"><a href="#footnote-1-source">1</a></sup> Footnote one.</p>

<p><sup id="footnote-2"><a href="#footnote-2-source">2</a></sup> Footnote two.</p>

</div>
`,
		transformFootnotes(`
<p>This is a reference [1] to a footnote [2].</p>

<p>Not footnote: KEYS[1].</p>

<p>[1] Footnote one.</p>

<p>[2] Footnote two.</p>
`,
			nil,
		),
	)

	// Without links
	assert.Equal(t, `
<p>This is a reference <sup><strong>1</strong></sup> to a footnote <sup><strong>2</strong></sup>.</p>

<p>Not footnote: KEYS[1].</p>


<div id="footnotes">
  <p><sup><strong>1</strong></sup> Footnote one.</p>

<p><sup><strong>2</strong></sup> Footnote two.</p>

</div>
`,
		transformFootnotes(`
<p>This is a reference [1] to a footnote [2].</p>

<p>Not footnote: KEYS[1].</p>

<p>[1] Footnote one.</p>

<p>[2] Footnote two.</p>
`,
			&RenderOptions{NoFootnoteLinks: true},
		),
	)
}

func TestTransformHeaders(t *testing.T) {
	assert.Equal(t, `
<h2 id="intro"><a href="#intro">Introduction</a></h2>

Intro here.

<h2 id="section-1"><a href="#section-1">Body</a></h2>

<h3 id="article"><a href="#article">Article</a></h3>

Article one.

<h3 id="sub"><a href="#sub">Subsection</a></h3>

More content.

<h3 id="article-1"><a href="#article-1">Article</a></h3>

Article two.

<h3 id="section-5"><a href="#section-5">Subsection</a></h3>

More content.

<h2 id="conclusion"><a href="#conclusion">Conclusion</a></h2>

Conclusion.
`,
		transformHeaders(`
## Introduction (#intro)

Intro here.

## Body

### Article (#article)

Article one.

### Subsection (#sub)

More content.

### Article (#article)

Article two.

### Subsection

More content.

## Conclusion (#conclusion)

Conclusion.
`,
			nil,
		),
	)

	assert.Equal(t, `
<h2>Introduction</h2>
`,
		transformHeaders(`
## Introduction (#intro)
`,
			&RenderOptions{NoHeaderLinks: true},
		),
	)
}

func TestTransformImagesToRetina(t *testing.T) {
	assert.Equal(t,
		`<img src="/assets/hello.jpg" srcset="/assets/hello@2x.jpg 2x, /assets/hello.jpg 1x">`,
		transformImagesToRetina(`<img src="/assets/hello.jpg">`, nil),
	)

	// No srcset is inserted for resolution agnostic SVGs.
	assert.Equal(t,
		`<img src="/assets/hello.svg">`,
		transformImagesToRetina(`<img src="/assets/hello.svg">`, nil),
	)

	// Make sure transformation works with other attributes in the <img> tag (I
	// previously introduced a bug relating to this).
	assert.Equal(t,
		`<img src="/assets/hello.svg" class="overflowing">`,
		transformImagesToRetina(`<img src="/assets/hello.svg" class="overflowing">`, nil),
	)

	// No replacement when we've explicitly requested no retina conversion
	assert.Equal(t,
		`<img src="/assets/hello.jpg">`,
		transformImagesToRetina(
			`<img src="/assets/hello.jpg">`,
			&RenderOptions{NoRetina: true},
		),
	)
}

func TestTransformImagesToAbsoluteURLs(t *testing.T) {
	assert.Equal(t,
		`<img src="https://brandur.org/assets/hello.jpg">`,
		transformImagesToAbsoluteURLs(
			`<img src="/assets/hello.jpg">`,
			&RenderOptions{AbsoluteURL: "https://brandur.org"},
		),
	)

	// URLs that are already absolute are left alone.
	assert.Equal(t,
		`<img src="https://example.com/assets/hello.jpg">`,
		transformImagesToAbsoluteURLs(
			`<img src="https://example.com/assets/hello.jpg">`,
			&RenderOptions{AbsoluteURL: "https://brandur.org"},
		),
	)

	// Should pass through if options are nil.
	assert.Equal(t,
		`<img src="/assets/hello.jpg">`,
		transformImagesToAbsoluteURLs(
			`<img src="/assets/hello.jpg">`,
			nil,
		),
	)
}

func TestTransformLinksToNoFollow(t *testing.T) {
	assert.Equal(t,
		`<a href="https://example.com" rel="nofollow">Example</a>`+
			`<span class="hello">Hello</span>`,
		transformLinksToNoFollow(
			`<a href="https://example.com">Example</a>`+
				`<span class="hello">Hello</span>`,
			&RenderOptions{NoFollow: true},
		),
	)

	// URLs that are relative should be left alone.
	assert.Equal(t,
		`<a href="/relative">Relative link</a>`,
		transformImagesToAbsoluteURLs(
			`<a href="/relative">Relative link</a>`,
			&RenderOptions{NoFollow: true},
		),
	)

	// Should pass through if options are nil.
	assert.Equal(t,
		`<a href="https://example.com">Example</a>`,
		transformImagesToAbsoluteURLs(
			`<a href="https://example.com">Example</a>`,
			nil,
		),
	)
}
