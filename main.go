package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/helpers"
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
			log.Println(err)
		}
	}

	dataReader := bytes.NewReader(state.Data)

	// Save mail in maildir
	filename, err := mailDir.CreateMail(dataReader)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Mail written to file: " + filename)
	}
}

func mail(state *mta.State) {
	log.Printf("From: %s\n", state.From.Address)
	log.Printf("To: ")
	for i, to := range state.To {
		log.Printf("%s", to.Address)
		if i != len(state.To)-1 {
			log.Printf(",")
		}
	}
	log.Printf("\nCONTENT_START:\n")
	log.Printf("%s\n", string(state.Data))
	log.Printf("CONTENT_END\n\n\n\n")
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
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	fmt.Println("GoPistolet at your service!")

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
		log.Println(err)
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
		log.Println(err)
	}
}
