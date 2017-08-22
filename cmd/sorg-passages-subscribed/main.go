package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joeshaw/envdecode"
	"gopkg.in/mailgun/mailgun-go.v1"
)

const (
	mailDomain  = "list.brandur.org"
	fromAddress = "Brandur <brandur@brandur.org>"
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

func renderAndSend(records [][]string, live bool) error {
	mg := mailgun.NewMailgun(mailDomain, conf.MailgunAPIKey, "")

	for _, record := range records {
		// skip empty lines
		if len(record) == 0 {
			continue
		}

		if len(record) < 2 {
			return fmt.Errorf("Record less than 2-width: %v", record)
		}

		name := record[0]
		recipient := record[1]

		subject := "I'm subscribing you to my newsletter"
		body := fmt.Sprintf(`%s,

One thing I've realized over the last few years is that I
don't do a very good job of keeping in touch with people.
On top of that, I also don't do much with social media, so
it's not even possible to passively stay appraised of what
I'm doing.

On a recent trip I was inspired to try and address this,
and to that end I'm trying something new: a newsletter.
Each issue will be a small compilation of stories and
thoughts. The bursts will be pretty infrequent; I'll send
out roughly as often as we get to witness a total solar
eclipse.

(This email is not the newsletter, but) I'm about to send
its first issue, and I've added you to the list to receive
it. If you don't want it, either reply to me here saying
so, or wait until I send it and click the very conspicuous
unsubscribe link (I won't get notified on an unsubscribe,
and even if I did, I wouldn't take it personally).

Thanks, and I hope you like it!

Brandur`,
			name,
		)

		if live {
			message := mailgun.NewMessage(fromAddress, subject, body, recipient)
			resp, _, err := mg.Send(message)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf(`Sent to: %s (response: "%s")`, recipient, resp)
		} else {
			fmt.Printf("To: %v <%v>\n", name, recipient)
			fmt.Printf("Subject: %v\n\n", subject)
			fmt.Printf("%v\n---\n\n", body)
		}
	}

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %v [-live] <recipient_file>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	live := flag.Bool("live", false,
		"Send to list (as opposed to dry run)")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
	}

	err := envdecode.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	err = renderAndSend(records, *live)
	if err != nil {
		log.Fatal(err)
	}
}
