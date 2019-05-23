package scommon

import (
	"fmt"
	"os"
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
