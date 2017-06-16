[![GoDoc](https://godoc.org/github.com/JohannesEbke/go-imap-sync?status.svg)](https://godoc.org/github.com/JohannesEbke/go-imap-sync)

# go-imap-sync
Library and command-line tool to sync the contents of an IMAP folder to disk, using https://github.com/emersion/go-imap.

## What does it do?
imapsync downloads all emails from an IMAP mailbox, each email into a plain text file. The names of the files are
derived from the email message ID. This enables imapsync to avoid downloading emails twice.

## Usage from the Command Line
```
go get github.com/JohannesEbke/go-imap-sync/cmd/go-imap-sync
go-imap-sync -server mail.example.com:993 -username me -mailbox INBOX
```
You will be prompted for your password. If you use the program in scripts, you can also set the `IMAP\_PASSWORD`
environment variable.

## Usage as a Library
```
err := imapsync.Sync("mail.example.com:993", "me", "hunter2", "INBOX", "/home/me/mails")
```

## Why?
I've encountered the problem of acquiring incoming email from IMAP now in two different projects,
https://github.com/martinhoefling/go-dmarc-report and https://github.com/TNG/openpgp-validation-server,
and decided to factor out the common parts.

## Roadmap
One possible additional mode of operation is to continuously monitor an IMAP mailbox, download new mails as they
appear and notify the caller of the new emails, e.g. via a channel.
