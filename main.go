package main

import (
	"bytes"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gopistolet/mta"
	"github.com/sloonz/go-maildir"
)

var mailDir *maildir.Maildir

func handleMailDir(state *mta.State) {
	err := errors.New("")

	// Open maildir if it's not yet open
	if mailDir == nil {

		// Open a maildir. If it does not exist, create it.
		mailDir, err = maildir.New("./maildir", true)
		if err != nil {
			log.Errorln(err)
		}
	}

	dataReader := bytes.NewReader(state.Data)

	// Save mail in maildir
	filename, err := mailDir.CreateMail(dataReader)
	if err != nil {
		//log.Println(err)
		log.WithFields(log.Fields{
			"SessionId": state.SessionId.String(),
		}).Error(err)
	} else {
		//log.Println("Maildir: mail written to file: " + filename)
		log.WithFields(log.Fields{
			"SessionId": state.SessionId.String(),
		}).Info("Maildir: mail written to file: " + filename)
	}
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

type Chain struct {
	handlers []mta.Handler
}

func (c *Chain) HandleMail(state *mta.State) {
	for _, handler := range c.handlers {
		handler.HandleMail(state)
	}
}

func main() {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	log.Println("GoPistolet at your service!")

	// Default config
	c := mta.Config{
		Hostname: "localhost",
		Port:     25,
		TlsCert:  "ssl/server.crt",
		TlsKey:   "ssl/server.key",
	}

	// Load config from JSON file
	err := helpers.DecodeFile("config.json", &c)
	if err != nil {
		log.Warnln(err, "- Using default configuration instead.")
	}

	mta := mta.NewDefault(c,
		&Chain{handlers: []mta.Handler{
			mta.HandlerFunc(mail),
			mta.HandlerFunc(handleMailDir),
		}})
	go func() {
		<-sigc
		mta.Stop()
	}()
	err = mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
