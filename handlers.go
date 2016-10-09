package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gospf"
	"github.com/gopistolet/gospf/dns"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
	"github.com/sloonz/go-maildir"
)

var mailQueue = make(chan smtp.State)

func handleQueue(state *smtp.State) {
	// Save mail to disk
	save(state)

	// Put mail in mail queue
	mailQueue <- (*state)
}

func MailQueueWorker(q chan smtp.State, handler mta.Handler) {

	for {
		state := <-q

		// Handle mail
		handler.HandleMail(&state)

		// Remove mail from disk
		delete(&state)

	}

}

type Chain struct {
	handlers []mta.Handler
}

func (c *Chain) HandleMail(state *smtp.State) {
	for _, handler := range c.handlers {
		handler.HandleMail(state)
	}
}

func handleMailStore(state *smtp.State) {

}

var mailDir *maildir.Maildir

func handleMailDir(state *smtp.State) {
	err := errors.New("")

	// Open maildir if it's not yet open
	if mailDir == nil {

		// Open a maildir. If it does not exist, create it.
		mailDir, err = maildir.New("./maildir", true)
		if err != nil {
			log.Errorf("Could not open maildir: %v", err)
			return
		}
	}

	dataReader := bytes.NewReader(state.Data)

	// Save mail in maildir
	filename, err := mailDir.CreateMail(dataReader)
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Error(err)
	} else {
		//log.Println("Maildir: mail written to file: " + filename)
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Info("Maildir: mail written to file: " + filename)
	}
}

func handleSPF(state *smtp.State) {
	// create SPF instance
	spf, err := gospf.New(state.From.GetDomain(), &dns.GoSPFDNS{})
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Infof("Could not create spf: %v", err)
		return
	}

	// check the given IP on that instance
	check, err := spf.CheckIP(state.Ip.String())
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Errorf("Error while checking ip in spf: %v", err)
		return
	}

	log.WithFields(log.Fields{
		"Ip":     state.Ip.String(),
		"Domain": state.From.GetDomain(),
	}).Info("SPF returned " + check)

	// write Authentication-Results header
	// TODO: need value from config here...
	//
	// header field is defined in RFC 5451 section 2.2
	// Authentication-Results: receiver.example.org; spf=pass smtp.mailfrom=example.com;
	headerField := fmt.Sprintf("Authentication-Results: %s; spf=%s smtp.mailfrom=%s;\r\n", hostname, strings.ToLower(check), state.From.GetDomain())
	state.Data = append([]byte(headerField), state.Data...)

}

func headerReceived(state *smtp.State) {

	/*
		RFC 2076 3.2 Trace information

		    Trace of MTAs which a message has passed.


		RFC 5322 3.6.7.

		    received        =   "Received:" *received-token ";" date-time CRLF

		    received-token  =   word / angle-addr / addr-spec / domain


		Example:

		    Received: from mail.example.com (192.168.0.10) by some.mail.server.example.com (192.168.0.11) with Microsoft SMTP Server id 14.3.319.2; Wed, 5 Oct 2016 14:57:46 +0200
	*/
	date := time.Now().Format(time.RFC1123Z) // date-time in RFC 5322 is like RFC 1123Z
	headerField := fmt.Sprintf("Received: from %s (%s) by %s (%s) with GoPistolet; %s\r\n", state.Hostname, state.Ip, c.Hostname, c.Ip, date)
	state.Data = append([]byte(headerField), state.Data...)

	// TODO: 'by IP' is not necessarily set in config

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
		"Hostname":  state.Hostname,
	}).Debug("Added 'received' header: '", headerField, "'")
}

func fileNameForState(state *smtp.State) (s string) {
	s += state.SessionId.String()
	s += "." + state.From.String()
	s += ".json"
	return
}

// Save mails to disk, since we are responsible for the message do be delivered
func save(state *smtp.State) {

	filename := "mailstore/" + fileNameForState(state)

	err := helpers.EncodeFile(filename, state)
	if err != nil {
		log.Fatal("Couldn't save mail to disk: ", err.Error())
	}

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
	}).Debug("Serialized mail to disk: ", filename)

}

func delete(state *smtp.State) {
	filename := "mailstore/" + fileNameForState(state)
	err := os.Remove(filename)
	if err != nil {
		log.Warnln("Couldn't save mail to disk: ", err.Error())
		return
	}

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
	}).Debug("Removed temp mail from disk: ", filename)
}
