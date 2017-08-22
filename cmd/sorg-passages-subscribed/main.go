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
		if len(record) < 2 {
			return fmt.Errorf("Record less than 2-width: %v", record)
		}

		name := record[0]
		recipient := record[1]

		subject := "I'm subscribing you to a mailing list"
		body := fmt.Sprintf(`Hi %s,

I'm subscribing you to a mailing list.

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

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	err = renderAndSend(records, *live)
	if err != nil {
		log.Fatal(err)
	}
}
