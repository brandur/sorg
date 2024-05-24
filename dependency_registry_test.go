package main

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyRegistryParseGoTemplate(t *testing.T) {
	dependencies := NewDependencyRegistry()
	emptyTmpl := template.New("base_empty")

	// Use some preexistting template for simplicity.
	{
		_, dependencies, err := dependencies.parseGoTemplate(template.Must(emptyTmpl.Clone()), "layouts/main.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"layouts/main.tmpl.html",
			"views/_tailwind_stylesheets.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_dark_mode_js.tmpl.html",
			"views/_analytics_js.tmpl.html",
			"views/_shiki_js.tmpl.html",
		}, dependencies)
	}

	{
		_, dependencies, err := dependencies.parseGoTemplate(template.Must(emptyTmpl.Clone()), "layouts/pages/belize.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"layouts/pages/belize.tmpl.html",
			"layouts/main.tmpl.html",
			"views/_tailwind_stylesheets.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_dark_mode_js.tmpl.html",
			"views/_analytics_js.tmpl.html",
			"views/_shiki_js.tmpl.html",
		}, dependencies)
	}

	{
		_, dependencies, err := dependencies.parseGoTemplate(template.Must(emptyTmpl.Clone()), "pages/belize/01.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"pages/belize/01.tmpl.html",
			"layouts/pages/belize.tmpl.html",
			"layouts/main.tmpl.html",
			"views/_tailwind_stylesheets.tmpl.html",
			"views/_twitter.tmpl.html",
			"views/_dark_mode_js.tmpl.html",
			"views/_analytics_js.tmpl.html",
			"views/_shiki_js.tmpl.html",
		}, dependencies)
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
