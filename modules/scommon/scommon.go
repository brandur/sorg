package scommon

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	"github.com/brandur/modulir/modules/mtemplate"
	"github.com/brandur/modulir/modules/mtemplatemd"
	"github.com/brandur/sorg/modules/stemplate"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Constants
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	// AtomAuthorName is the name of the author to include in Atom feeds.
	AtomAuthorName = "Brandur Leach"

	// AtomTag is a stable constant to use in Atom tags.
	AtomTag = "brandur.org"

	// DataDir is where various TOML files for quantified self statistics
	// reside. These are pulled from another project which updates them
	// automatically.
	DataDir = "./data"

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts"

	// MainLayout is the site's main layout.
	MainLayout = LayoutsDir + "/main.ace"

	// NanoglyphsLayout is the layout for a Nanoglyph issue (an email
	// newsletter).
	NanoglyphsLayout = LayoutsDir + "/nanoglyphs.ace"

	// PassagesLayout is the layout for a Passages & Glass issue (an email
	// newsletter).
	PassagesLayout = LayoutsDir + "/passages.ace"

	// TempDir is a temporary directory used to download images that will be
	// processed and such.
	TempDir = "./tmp"

	// TitleSuffix is the suffix to add to the end of page and Atom titles.
	TitleSuffix = " — brandur.org"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views"
)

// TwitterInfo is some HTML that includes a Twitter link which can be appended
// to the publishing info of various content.
const TwitterInfo = template.HTML(`<p>Find me on Twitter at ` +
	`<strong><a href="https://twitter.com/brandur">@brandur</a></strong>.</p>`)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// HTMLTemplateFuncMap is a function map of template helpers which is the
// combined version of the maps from ftemplate, mtemplate, and mtemplatemd.
var HTMLTemplateFuncMap template.FuncMap = mtemplate.CombineFuncMaps(
	stemplate.FuncMap,
	mtemplate.FuncMap,
	mtemplatemd.FuncMap,
)

// TextTemplateFuncMap is a combined set of template helpers for text
// templates.
var TextTemplateFuncMap texttemplate.FuncMap = mtemplate.HTMLFuncMapToText(HTMLTemplateFuncMap)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

// ExitWithError prints the given error to stderr and exits with a status of 1.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

// ExtractSlug gets a slug for the given filename by using its basename
// stripped of file extension.
func ExtractSlug(source string) string {
	return strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
}

// IsDraft does really simplistic detection on whether the given source is a
// draft by looking whether the name "drafts" is in its parent directory's
// name.
func IsDraft(source string) bool {
	return strings.Contains(filepath.Base(filepath.Dir(source)), "drafts")
}
