package main

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/aymerick/douceur/inliner"
	"github.com/brandur/modulir"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/spassages"
	"github.com/brandur/sorg/modules/stemplate"
	"github.com/yosssi/ace"
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

func sendPassages(c *modulir.Context, source string, live, staging bool) {
	if err := renderAndSend(c, source, live, staging); err != nil {
		scommon.ExitWithError(err)
	}
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
	mailDomain         = "list.brandur.org"
	fromAddress        = "Brandur <" + listAddress + ">"
	listAddress        = "passages@" + mailDomain
	listStagingAddress = "passages-staging@" + mailDomain
	replyToAddress     = "brandur@brandur.org"
	testAddress        = replyToAddress
)

func renderAndSend(c *modulir.Context, source string, live, staging bool) error {
	if conf.MailgunAPIKey == "" {
		scommon.ExitWithError(fmt.Errorf(
			"MAILGUN_API_KEY must be configured in the environment"))
	}

	dir := filepath.Dir(source)
	name := filepath.Base(source)

	passage, err := spassages.Render(c, dir, name, conf.AbsoluteURL, true)
	if err != nil {
		return err
	}

	locals := map[string]interface{}{
		"InEmail": true,
		"Passage": passage,
		"Title":   passage.Title,
	}

	template, err := ace.Load(
		scommon.PassageLayout,
		scommon.ViewsDir+"/passages/show",
		&ace.Options{FuncMap: stemplate.FuncMap})
	if err != nil {
		return err
	}

	var b bytes.Buffer

	writer := bufio.NewWriter(&b)

	err = template.Execute(writer, locals)
	if err != nil {
		return err
	}

	writer.Flush()

	html, err := inliner.Inline(b.String())
	if err != nil {
		return err
	}

	var recipient string
	if live {
		recipient = listAddress
	} else if staging {
		recipient = listStagingAddress
	} else {
		recipient = testAddress
	}

	mg := mailgun.NewMailgun(mailDomain, conf.MailgunAPIKey, "")

	subject := fmt.Sprintf("Passages & Glass %s — %s",
		passage.Issue, passage.Title)

	message := mailgun.NewMessage(
		fromAddress,
		subject,
		passage.ContentRaw,
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
