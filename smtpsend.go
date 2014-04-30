package main

import (
	"flag"
	"fmt"
	"github.com/zerklabs/libsmtp"
	"log"
	"strings"
	"time"
)

var (
	attachment = flag.String("attachment", "", "Include the attachment with the message")
	body       = flag.String("body", "This is a test message", "Set the body of the message")
	from       = flag.String("from", "", "Set the mail sender")
	port       = flag.Int("port", 25, "Set the SMTP server port. Default is 25")
	server     = flag.String("server", "", "Set the SMTP server")
	subject    = flag.String("subject", fmt.Sprintf("smtpsend test - %s", time.Now()), "Set the mail subject")
	useTls     = flag.Bool("tls", false, "If given, will try and send the message with STARTTLS")
	to         = flag.String("to", "", "Set the mail recipient(s). Separate multiple entries with commas")
)

func main() {
	flag.Parse()

	client, err := libsmtp.New(*server, *port, *from, strings.Split(*to, ","), *useTls)

	if err != nil {
		log.Fatal(err)
	}

	client.Subject(*subject)
	client.Body(*body)

	if len(*attachment) > 0 {
		if err = client.AddAttachment(*attachment); err != nil {
			log.Fatal(err)
		}
	}

	if err = client.Send(); err != nil {
		log.Fatal(err)
	}

	log.Println("Message Sent")
}
