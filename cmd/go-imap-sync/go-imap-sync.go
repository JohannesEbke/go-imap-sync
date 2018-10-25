// go-imap-sync provides a simple command line tool to download emails from an IMAP mailbox. Each email is saved as a
// plain text file (per default in the messages/ subdirectory). Emails are only downloaded once if run repeatedly.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/JohannesEbke/go-imap-sync"
	"github.com/howeyc/gopass"
)

func getPassword(username, server string) (password string) {
	password = os.Getenv("IMAP_PASSWORD")

	if password == "" {
		log.Printf("Enter IMAP Password for %v on %v: ", username, server)
		passwordBytes, err := gopass.GetPasswd()
		if err != nil {
			panic(err)
		}
		password = string(passwordBytes)
	}
	return
}

func main() {
	var server, username, mailbox, emailDir string
	var disableTls bool
	flag.StringVar(&server, "server", "", "sync from this mail server and port (e.g. mail.example.com:993)")
	flag.StringVar(&username, "username", "", "username for logging into the mail server")
	flag.StringVar(&mailbox, "mailbox", "", "mailbox to read messages from (typically INBOX or INBOX/subfolder)")
	flag.StringVar(&emailDir, "messagesDir", "messages", "local directory to save messages in")
	flag.BoolVar(&disableTls, "disableTls", false, "optionally disable TLS for IMAP")
	flag.Parse()

	if server == "" {
		log.Println("go-imap-sync copies emails from an IMAP mailbox to your computer. Usage:")
		flag.PrintDefaults()
		log.Fatal("Required parameters not found.")
	}

	password := getPassword(username, server)

	_, err := imapsync.Sync(server, username, password, mailbox, emailDir, disableTls)
	if err != nil {
		log.Fatal(err)
	}
}
