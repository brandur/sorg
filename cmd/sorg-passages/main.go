package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aymerick/douceur/inliner"
	"github.com/brandur/sorg"
	"github.com/brandur/sorg/passages"
	"github.com/brandur/sorg/templatehelpers"
	"github.com/joeshaw/envdecode"
	"github.com/yosssi/ace"
	"gopkg.in/mailgun/mailgun-go.v1"
)

const (
	mailDomain         = "list.brandur.org"
	fromAddress        = "Brandur <" + listAddress + ">"
	listAddress        = "passages@" + mailDomain
	listStagingAddress = "passages-staging@" + mailDomain
	replyToAddress     = "brandur@brandur.org"
	testAddress        = replyToAddress
)

// Conf contains configuration information for the command. It's extracted from
// environment variables.
type Conf struct {
	// MailgunAPIKey is a key for Mailgun used to send email.
	MailgunAPIKey string `env:"MAILGUN_API_KEY,required"`
}

// Left as a global for now for the sake of convenience, but it's not used in
// very many places and can probably be refactored as a local if desired.
var conf Conf

func renderAndSend(path string, live, staging bool) error {
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	passage, err := passages.Compile(dir, name, false, true)
	if err != nil {
		return err
	}

	locals := map[string]interface{}{
		"InEmail": true,
		"Passage": passage,
		"Title":   passage.Title,
	}

	template, err := ace.Load(
		sorg.PassageLayout,
		sorg.ViewsDir+"/passages/show",
		&ace.Options{FuncMap: templatehelpers.FuncMap})
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
		log.Fatal(err)
	}
	log.Printf(`Sent to: %s (response: "%s")`, recipient, resp)

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [-live] [-staging] <source_file>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	live := flag.Bool("live", false,
		"Send to list (as opposed to dry run)")
	staging := flag.Bool("staging", false,
		"Send to staging list (as opposed to dry run)")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
	}

	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	err = renderAndSend(flag.Arg(0), *live, *staging)
	if err != nil {
		log.Fatal(err)
	}
}
