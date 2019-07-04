package scommon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts"

	// MainLayout is the site's main layout.
	MainLayout = LayoutsDir + "/main.ace"

	// NanoglyphLayout is the layout for a Nanoglyph issue (an email
	// newsletter).
	NanoglyphLayout = LayoutsDir + "/nanoglyphs.ace"

	// PassageLayout is the layout for a Passages & Glass issue (an email
	// newsletter).
	PassageLayout = LayoutsDir + "/passages.ace"

	// TempDir is a temporary directory used to download images that will be
	// processed and such.
	TempDir = "./tmp"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views"
)

// TwitterInfo is some HTML that includes a Twitter link which can be appended
// to the publishing info of various content.
const TwitterInfo = `<p>Find me on Twitter at ` +
	`<strong><a href="https://twitter.com/brandur" class="twitter-icon-nav">@brandur</a></strong>.</p>`

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
