package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gospf"
	"github.com/gopistolet/gospf/dns"
	"github.com/gopistolet/smtp/mta"
	"github.com/sloonz/go-maildir"
)

var mailQueue = make(chan mta.State)

func handleQueue(state *mta.State) {
	// Save mail to disk
	save(state)

	// Put mail in mail queue
	mailQueue <- (*state)
}

func MailQueueWorker(q chan mta.State, handler mta.Handler) {

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

func (c *Chain) HandleMail(state *mta.State) {
	for _, handler := range c.handlers {
		handler.HandleMail(state)
	}
}

var mailDir *maildir.Maildir

func handleMailDir(state *mta.State) {
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

func handleSPF(state *mta.State) {
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
	headerField := fmt.Sprintf("Authentication-Result: %s; spf=%s smtp.mailfrom=%s;\r\n", hostname, strings.ToLower(check), state.From.GetDomain())
	state.Data = append([]byte(headerField), state.Data...)

}

func mail(state *mta.State) {
	log.Debugf("From: %s\n", state.From.Address)
	log.Debugf("To: ")
	for i, to := range state.To {
		log.Printf("%s", to.Address)
		if i != len(state.To)-1 {
			log.Printf(",")
		}
	}
	log.Debugf("CONTENT_START:\n")
	log.Debugf("%s\n", string(state.Data))
	log.Debugf("CONTENT_END\n")
}

func fileNameForState(state *mta.State) (s string) {
	s += state.SessionId.String()
	s += "." + state.From.String()
	s += ".json"
	return
}

// Save mails to disk, since we are responsible for the message do be delivered
func save(state *mta.State) {

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

func delete(state *mta.State) {
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
