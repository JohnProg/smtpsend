package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/zerklabs/auburn"
	"io/ioutil"
	"log"
	"net/smtp"
	"path/filepath"
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

	if len(*server) == 0 {
		log.Fatal("SMTP server required")
	}

	if len(*from) == 0 {
		log.Fatal("Mail sender required")
	}

	if len(*to) == 0 {
		log.Fatal("Mail recipient(s) required")
	}

	buildMailMessage()
}

func buildMailMessage() {
	boundary := auburn.RandomBase36()
	buf := bytes.NewBuffer(nil)

	buf.WriteString(fmt.Sprintf("From: %s\n", *from))
	buf.WriteString(fmt.Sprintf("To: %s\n", *to))
	buf.WriteString(fmt.Sprintf("Subject: %s\n", *subject))
	buf.WriteString("MIME-version: 1.0;\n")

	if len(*attachment) > 0 {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	}

	buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\";\n\n")
	buf.WriteString(*body)

	if len(*attachment) > 0 {
		attachmentName := filepath.Base(*attachment)
		b, err := ioutil.ReadFile(*attachment)

		if err != nil {
			log.Fatalf("Problem reading the given attachment:\n\t%s", err)
		}

		encodedLen := base64.StdEncoding.EncodedLen(len(b))
		encodedAttachment := make([]byte, encodedLen)
		base64.StdEncoding.Encode(encodedAttachment, b)

		buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: application/octet-stream; name=\"%s\"\n", attachmentName))
		buf.WriteString(fmt.Sprintf("Content-Description: %s\n", attachmentName))
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"; size=%d\n", attachmentName, encodedLen))
		buf.WriteString("Content-Transfer-Encoding: base64\n\n")

		buf.Write(encodedAttachment)
		buf.WriteString(fmt.Sprintf("\n--%s--", boundary))
	}

	smtpUri := fmt.Sprintf("%s:%d", *server, *port)

	c, err := smtp.Dial(smtpUri)

	if err != nil {
		log.Fatalf("Error creating SMTP connection: %s", err)
	}

	if *useTls {
		// check if TLS is supported
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err = c.StartTLS(&tls.Config{InsecureSkipVerify: true, ServerName: *server}); err != nil {
				c.Reset()
				c.Quit()

				log.Fatalf("Failed to establish TLS session: %s", err)
			}

			log.Println("TLS negotiated, sending over an encrypted channel")
		} else {
			log.Println("Server doesn't support TLS.. Sending over an unencrypted channel")
		}
	}

	// set the from addr
	if err = c.Mail(*from); err != nil {
		c.Reset()
		c.Quit()

		log.Fatalf("Failed to set the From address: %s", err)
	}

	// add the recipients
	for _, v := range strings.Split(*to, ",") {
		if err = c.Rcpt(v); err != nil {
			c.Reset()
			c.Quit()

			log.Fatalf("Failed to set a recipient: %s", err)
		}
	}

	w, err := c.Data()

	if err != nil {
		c.Reset()
		c.Quit()

		log.Fatalf("Failed to issue DATA command: %s", err)
	}

	_, err = w.Write(buf.Bytes())

	if err != nil {
		c.Reset()
		c.Quit()

		log.Fatalf("Failed to write DATA: %s", err)
	}

	if err = w.Close(); err != nil {
		c.Reset()
		c.Quit()

		log.Fatalf("Failed to close the DATA stream: %s", err)
	}

	c.Quit()

	log.Println("Message Sent")
}
