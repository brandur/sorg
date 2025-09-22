package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aymerick/douceur/inliner"
	"github.com/mailgun/mailgun-go/v4"
	"golang.org/x/xerrors"

	"github.com/brandur/modulir"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/snewsletter"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

func sendNewsletter(c *modulir.Context, source string, live, staging bool) {
	ctx := context.Background()

	if err := renderAndSend(ctx, c, source, live, staging); err != nil {
		scommon.ExitWithError(err)
	}
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Contains some metadata for a newsletter around email addresses, paths, etc.
type newsletterInfo struct {
	ContentDir         string
	ContentDirDrafts   string
	Layout             string
	MailAddress        string
	MailAddressStaging string
	TitleFormat        string
	View               string
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

const (
	mailDomain     = "list.brandur.org"
	replyToAddress = "brandur@brandur.org"
	testAddress    = replyToAddress
)

var (
	// Metadata for the Nanoglyph newsletter.
	newsletterInfoNanoglyphs = &newsletterInfo{
		ContentDir:         "./content/nanoglyphs",
		ContentDirDrafts:   "./content/nanoglyphs-drafts",
		Layout:             scommon.NanoglyphsLayout,
		MailAddress:        "nanoglyph@" + mailDomain,
		MailAddressStaging: "nanoglyph-staging@" + mailDomain,
		TitleFormat:        "ⓝ Nanoglyph %s — %s",
		View:               scommon.ViewsDir + "/nanoglyphs/show.tmpl.html",
	}

	// Metadata for the Passages & Glass newsletter.
	newsletterInfoPassages = &newsletterInfo{
		ContentDir:         "./content/passages",
		ContentDirDrafts:   "./content/passages-drafts",
		Layout:             scommon.PassagesLayout,
		MailAddress:        "passages@" + mailDomain,
		MailAddressStaging: "passages-staging@" + mailDomain,
		TitleFormat:        "Passages & Glass %s — %s",
		View:               scommon.ViewsDir + "/passages/show.tmpl.html",
	}

	// All infos combined into a slice.
	newsletterInfos = []*newsletterInfo{
		newsletterInfoNanoglyphs,
		newsletterInfoPassages,
	}
)

// Matches a known newsletter based on the path to the source file. The second
// return argument is true if the source is a draft.
func matchNewsletter(source string) (*newsletterInfo, bool) {
	source = filepath.Clean(source)

	for _, info := range newsletterInfos {
		{
			dir := filepath.Clean(info.ContentDirDrafts)
			if strings.HasPrefix(source, dir) {
				return info, true
			}
		}

		{
			dir := filepath.Clean(info.ContentDir)
			if strings.HasPrefix(source, dir) {
				return info, false
			}
		}
	}
	return nil, false
}

func renderAndSend(ctx context.Context, c *modulir.Context, source string, live, staging bool) error {
	if conf.MailgunAPIKey == "" {
		return xerrors.Errorf(
			"MAILGUN_API_KEY must be configured in the environment")
	}

	newsletterInfo, draft := matchNewsletter(source)
	if newsletterInfo == nil {
		return xerrors.Errorf("'%s' does not appear to be a known newsletter (check its path)",
			source)
	}

	if live && draft {
		return xerrors.Errorf("refusing to send a draft newsletter to a live list")
	}

	dir := filepath.Dir(source)
	name := filepath.Base(source)

	issue, err := snewsletter.Render(c, dir, name, conf.AbsoluteURL, true)
	if err != nil {
		return err
	}

	locals := map[string]interface{}{
		"InEmail":   true,
		"Issue":     issue,
		"Title":     issue.Title,
		"URLPrefix": conf.AbsoluteURL,
	}

	var (
		b            bytes.Buffer
		dependencies = NewDependencyRegistry()
	)

	err = dependencies.renderGoTemplateWriter(ctx, c, newsletterInfo.View, &b, locals)
	if err != nil {
		return err
	}

	html, err := inliner.Inline(b.String())
	if err != nil {
		return xerrors.Errorf("error inlining CSS: %w", err)
	}

	var recipient string
	switch {
	case live:
		recipient = newsletterInfo.MailAddress
	case staging:
		recipient = newsletterInfo.MailAddressStaging
	default:
		recipient = testAddress
	}

	mg := mailgun.NewMailgun(mailDomain, conf.MailgunAPIKey)

	fromAddress := "Brandur <" + newsletterInfo.MailAddress + ">"
	subject := fmt.Sprintf(newsletterInfo.TitleFormat,
		issue.Number, issue.Title)

	message := mg.NewMessage(
		fromAddress,
		subject,
		issue.ContentRaw,
		recipient)
	message.SetReplyTo(replyToAddress)
	message.SetHtml(html)

	resp, _, err := mg.Send(ctx, message)
	if err != nil {
		return xerrors.Errorf("error sending email: %w", err)
	}

	c.Log.Infof(`Sent to: %s (response: "%s")`, recipient, resp)
	return nil
}
