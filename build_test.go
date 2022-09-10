package main

import (
	"fmt"
	"html/template"
	"os"
	"testing"

	"github.com/joeshaw/envdecode"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}

func TestFindGoSubTemplates(t *testing.T) {
	require.Equal(t, []string{"layouts/main.tmpl.html"}, findGoSubTemplates(`{{template "layouts/main.tmpl.html" .}}`))
	require.Equal(t, []string{"layouts/main.tmpl.html"}, findGoSubTemplates(`{{template "layouts/main.tmpl.html" .}}`))
	require.Equal(t,
		[]string{"layouts/main.tmpl.html", "views/_other.tmpl.html"},
		findGoSubTemplates(`{{template "layouts/main.tmpl.html" .}}{{template "views/_other.tmpl.html" .}}`),
	)
	require.Equal(t, []string{}, findGoSubTemplates(`no templates here`))
	require.Equal(t, []string{}, findGoSubTemplates(``))
}

func TestParseGoTemplate(t *testing.T) {
	emptyTmpl := template.New("base_empty")

	// Use some preexistting template for simplicity.
	{
		_, dependencies, err := parseGoTemplate(template.Must(emptyTmpl.Clone()), "layouts/pages/main.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"layouts/pages/main.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_analytics.tmpl.html",
		}, dependencies)
	}

	{
		_, dependencies, err := parseGoTemplate(template.Must(emptyTmpl.Clone()), "layouts/pages/belize.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"layouts/pages/belize.tmpl.html",
			"layouts/pages/main.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_analytics.tmpl.html",
		}, dependencies)
	}

	{
		_, dependencies, err := parseGoTemplate(template.Must(emptyTmpl.Clone()), "pages/belize/01.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"pages/belize/01.tmpl.html",
			"layouts/pages/belize.tmpl.html",
			"layouts/pages/main.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_analytics.tmpl.html",
		}, dependencies)
	}
}

func TestPagePathKey(t *testing.T) {
	require.Equal(t, "about", pagePathKey("./pages/about.ace"))
	require.Equal(t, "about", pagePathKey("./pages-drafts/about.ace"))

	require.Equal(t, "deep/about", pagePathKey("./pages/deep/about.ace"))
	require.Equal(t, "deep/about", pagePathKey("./pages-drafts/deep/about.ace"))

	require.Equal(t, "really/deep/about", pagePathKey("./pages/really/deep/about.ace"))
	require.Equal(t, "really/deep/about", pagePathKey("./pages-drafts/really/deep/about.ace"))
}
