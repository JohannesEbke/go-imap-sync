// Package imapsync provides a single function to download an IMAP folder to a local directory, with each email
// in a plain text file. Emails are downloaded only once, even if the function is run repeatedly.
//
// A command line tool is available at https://github.com/JohannesEbke/go-imap-sync/cmd/go-imap-sync
package imapsync

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

type Result struct {
	ExistingEmails []string
	NewEmails      []string
}

// Sync downloads and saves all not-yet downloaded emails from the mailbox to the emailDir
func Sync(server, user, password, mailbox, emailDir string) (*Result, error) {
	err := os.MkdirAll(emailDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("Error creating email directory %v: %v", emailDir, err)
	}

	connection, err := connect(server, user, password)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to %v: %v", server, err)
	}

	defer func() {
		err2 := connection.Logout()
		if err2 != nil {
			log.Printf("Error on logout from %v: %v", server, err2)
		}
	}()

	_, err = connection.Select(mailbox, true)
	if err != nil {
		return nil, fmt.Errorf("Error opening mailbox '%v': %v", mailbox, err)
	}

	log.Printf("Listing all messages in %v...", mailbox)
	seqNumMessageIDMap := getMessageIDMap(connection)
	log.Printf("Found %v messages. Looking for existing downloaded messages...", len(seqNumMessageIDMap))
	messagesToFetch, toFetchCount := getMessagesToFetch(emailDir, seqNumMessageIDMap)
	log.Printf("Syncing %v missing messages...", toFetchCount)
	for _, messageSeqNr := range messagesToFetch {
		seqSet := new(imap.SeqSet)
		seqSet.AddNum(messageSeqNr)
		err2 := fetchMessages(connection, emailDir, seqSet)
		if err2 != nil {
			return nil, fmt.Errorf("Error fetching message '%v': %v", messageSeqNr, err2)
		}
		fmt.Println(".")
	}
	log.Printf("Finished syncing.")

	// Calculate Result structure
	isNew := make(map[uint32]bool)
	existingEmails := []string{}
	newEmails := []string{}
	for _, seqNum := range messagesToFetch {
		newEmails = append(newEmails, messageFileName(emailDir, seqNumMessageIDMap[seqNum]))
		isNew[seqNum] = true
	}
	for seqNum, messageId := range seqNumMessageIDMap {
		if !isNew[seqNum] {
			existingEmails = append(existingEmails, messageFileName(emailDir, messageId))
		}
	}
	return &Result{
		ExistingEmails: existingEmails,
		NewEmails:      newEmails,
	}, nil
}

// connect performs an interactive connection to the given IMAP server
func connect(server, username, password string) (*client.Client, error) {
	log.Printf("Connecting to %v...", server)
	c, err := client.DialTLS(server, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to %v.", server)

	if err := c.Login(username, password); err != nil {
		if err2 := c.Logout(); err2 != nil {
			return nil, fmt.Errorf("Error while logging in to %v: %v\n(followup error: %v)", server, err, err2)
		}
		return nil, fmt.Errorf("Error while logging in to %v: %v", server, err)
	}
	log.Printf("Logged in as user %v on %v.", username, server)
	return c, nil
}

// getMessageIDMap lists all messages in the current mailbox and returns a map of sequence numbers to email MessageID
func getMessageIDMap(c *client.Client) (emails map[uint32]string) {
	emails = make(map[uint32]string)
	// Get all messages
	seqset, err := imap.ParseSeqSet("1:*")
	if err != nil {
		log.Fatal(err)
	}
	messageChan := make(chan *imap.Message)
	go func() {
		if err := c.Fetch(seqset, []string{"ENVELOPE", "UID"}, messageChan); err != nil {
			log.Fatalf("Error fetching list of messages: %v", err)
		}
	}()
	for msg := range messageChan {
		emails[msg.SeqNum] = msg.Envelope.MessageId
	}
	return
}

// getMessagesToFetch determines which messages listed in the map should be downloaded from the server
func getMessagesToFetch(emailDir string, seqNumMessageIDMap map[uint32]string) (messagesToFetch []uint32, toFetchCount uint) {
	for seqNum, messageID := range seqNumMessageIDMap {
		exists, err := fileExists(messageFileName(emailDir, messageID))
		if err != nil {
			log.Fatal(err)
		}
		if !exists {
			messagesToFetch = append(messagesToFetch, seqNum)
			toFetchCount++
		}
	}
	return
}

// fetchMessages downloads the messages specified by the given SeqSet to the emailDir
func fetchMessages(connection *client.Client, emailDir string, messagesToFetch *imap.SeqSet) error {
	messageChan := make(chan *imap.Message)
	err := connection.Fetch(messagesToFetch, []string{"ENVELOPE", "BODY[]"}, messageChan)
	if err != nil {
		return err
	}
	for msg := range messageChan {
		body, err := ioutil.ReadAll(msg.GetBody("BODY[]"))
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(messageFileName(emailDir, msg.Envelope.MessageId), body, 0600)
		if err != nil {
			return err
		}
	}
	return nil
}

// sha512TruncatedHex returns a hex representation of the first 32 bytes of the SHA512 hash of the given string
func sha512TruncatedHex(messageID string) string {
	h := sha512.New()
	if _, err := io.WriteString(h, messageID); err != nil {
		log.Fatal(err)
	}
	b := h.Sum(nil)
	return hex.EncodeToString(b[0:31])
}

// messageFileName returns the target file name of the email with the given messageID
func messageFileName(emailDir, messageID string) string {
	return filepath.Join(emailDir, fmt.Sprintf("%s.eml", sha512TruncatedHex(messageID)))
}

// fileExists checks if the given path exists and can be Stat'd.
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
