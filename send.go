package main

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aymerick/douceur/inliner"
	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/mace"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/snewsletter"
	"github.com/brandur/sorg/modules/stemplate"
	"gopkg.in/mailgun/mailgun-go.v1"
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
	if err := renderAndSend(c, source, live, staging); err != nil {
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
		TitleFormat:        "Nanoglyph %s — %s",
		View:               scommon.ViewsDir + "/nanoglyphs/show.ace",
	}

	// Metadata for the Passages & Glass newsletter.
	newsletterInfoPassages = &newsletterInfo{
		ContentDir:         "./content/passages",
		ContentDirDrafts:   "./content/passages-drafts",
		Layout:             scommon.PassagesLayout,
		MailAddress:        "passages@" + mailDomain,
		MailAddressStaging: "passages-staging@" + mailDomain,
		TitleFormat:        "Passages & Glass %s — %s",
		View:               scommon.ViewsDir + "/passages/show.ace",
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

func renderAndSend(c *modulir.Context, source string, live, staging bool) error {
	if conf.MailgunAPIKey == "" {
		return fmt.Errorf(
			"MAILGUN_API_KEY must be configured in the environment")
	}

	newsletterInfo, draft := matchNewsletter(source)
	if newsletterInfo == nil {
		return fmt.Errorf("'%s' does not appear to be a known newsletter (check its path)",
			source)
	}

	if live && draft {
		return fmt.Errorf("refusing to send a draft newsletter to a live list")
	}

	dir := filepath.Dir(source)
	name := filepath.Base(source)

	issue, err := snewsletter.Render(c, dir, name, conf.AbsoluteURL, true)
	if err != nil {
		return err
	}

	locals := map[string]interface{}{
		"InEmail": true,
		"Issue":   issue,
		"Title":   issue.Title,
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	err = mace.Render(c, newsletterInfo.Layout, newsletterInfo.View,
		writer, stemplate.GetAceOptions(true), locals)

	writer.Flush()

	html, err := inliner.Inline(b.String())
	if err != nil {
		return err
	}

	var recipient string
	if live {
		recipient = newsletterInfo.MailAddress
	} else if staging {
		recipient = newsletterInfo.MailAddressStaging
	} else {
		recipient = testAddress
	}

	mg := mailgun.NewMailgun(mailDomain, conf.MailgunAPIKey, "")

	fromAddress := "Brandur <" + newsletterInfo.MailAddress + ">"
	subject := fmt.Sprintf(newsletterInfo.TitleFormat,
		issue.Number, issue.Title)

	message := mailgun.NewMessage(
		fromAddress,
		subject,
		issue.ContentRaw,
		recipient)
	message.SetReplyTo(replyToAddress)
	message.SetHtml(html)

	resp, _, err := mg.Send(message)
	if err != nil {
		return err
	}

	c.Log.Infof(`Sent to: %s (response: "%s")`, recipient, resp)
	return nil
}
