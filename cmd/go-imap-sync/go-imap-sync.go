package main

import (
	"flag"
	"fmt"
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
	flag.StringVar(&server, "server", "", "sync from this mail server and port (e.g. mail.example.com:993)")
	flag.StringVar(&username, "username", "", "username for logging into the mail server")
	flag.StringVar(&mailbox, "mailbox", "", "mailbox to read messages from (typically INBOX or INBOX/subfolder)")
	flag.StringVar(&emailDir, "messagesDir", "messages", "local directory to save messages in")
	flag.Parse()

	if server == "" {
		fmt.Println("go-imap-sync copies emails from an IMAP mailbox to your computer. Usage:")
		flag.PrintDefaults()
		return
	}

	password := getPassword(username, server)

	err := imapsync.Sync(server, username, password, mailbox, emailDir)
	if err != nil {
		log.Fatal(err)
	}
}
