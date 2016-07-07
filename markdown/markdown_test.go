package markdown

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", Render("**strong**"))
}

func TestRenderMarkdown(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", renderMarkdown("**strong**"))
}

func TestTransformCodeWithLanguagePrefix(t *testing.T) {
	assert.Equal(t,
		`<code class="language-ruby">`,
		transformCodeWithLanguagePrefix(`<code class="ruby">`),
	)
}

func TestTransformFigures(t *testing.T) {
	assert.Equal(t,
		`
<figure>
  <p><img src="fig-src"></p>
  <figcaption>fig-caption</figcaption>
</figure>
`,
		transformFigures(`!fig src="fig-src" caption="fig-caption"`),
	)
}
