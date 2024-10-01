package main

import (
	"html/template"
	"strings"
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

func TestFormatHTMLPreservingScripts(t *testing.T) {
	require.Equal(t, strings.TrimSpace(`
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>
      A title
    </title>
    <script type="module">
      const A_CONSTANT = "a constant";

      function setDocumentClasses(...classes) {
          THEME_CLASS_ALL.forEach((themeClass) => {
              if (classes.includes(themeClass)) {
                  document.documentElement.classList.add(themeClass)
              } else {
                  document.documentElement.classList.remove(themeClass)
              }
          })
      }
	</script>
  </head>
  <body>
    <script>const ONE_LINER = "one line constant";</script>
    <script>
      const A_CONSTANT = "a constant";

      function setDocumentClasses(...classes) {
          THEME_CLASS_ALL.forEach((themeClass) => {
              if (classes.includes(themeClass)) {
                  document.documentElement.classList.add(themeClass)
              } else {
                  document.documentElement.classList.remove(themeClass)
              }
          })
      }
	</script>
  </body>
</html>
	`), formatHTMLPreservingScripts(strings.TrimSpace(`
<!doctype html>

<html lang="en">
  <head>
  <meta charset="utf-8">
      <title>A title</title>
    <script type="module">
      const A_CONSTANT = "a constant";

      function setDocumentClasses(...classes) {
          THEME_CLASS_ALL.forEach((themeClass) => {
              if (classes.includes(themeClass)) {
                  document.documentElement.classList.add(themeClass)
              } else {
                  document.documentElement.classList.remove(themeClass)
              }
          })
      }
	</script>
  </head>
  <body>
    <script>const ONE_LINER = "one line constant";</script>
    <script>
      const A_CONSTANT = "a constant";

      function setDocumentClasses(...classes) {
          THEME_CLASS_ALL.forEach((themeClass) => {
              if (classes.includes(themeClass)) {
                  document.documentElement.classList.add(themeClass)
              } else {
                  document.documentElement.classList.remove(themeClass)
              }
          })
      }
	</script>
  </body>
</html>
`)))
}
