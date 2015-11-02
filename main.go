package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/gopistolet/mta"
)

func MailQueueWorker(q chan *mta.State, handler mta.Handler) {

	for {
		state := <-q
		log.Println("MailQueuWorker read state from channel:", state)
		handler.HandleMail(state)
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

var hostname string

func main() {

	log.SetLevel(log.DebugLevel)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	mailQueue := make(chan *mta.State)
	go MailQueueWorker(mailQueue, &Chain{handlers: []mta.Handler{
		mta.HandlerFunc(mail),
		mta.HandlerFunc(handleSPF),
		mta.HandlerFunc(handleMailDir),
	}})

	log.Println("GoPistolet at your service!")

	// Default config
	c := mta.Config{
		Hostname:  "localhost",
		Port:      25,
		TlsCert:   "ssl/server.crt",
		TlsKey:    "ssl/server.key",
		MailQueue: mailQueue,
	}

	// Load config from JSON file
	err := helpers.DecodeFile("config.json", &c)
	if err != nil {
		log.Warnln(err, "- Using default configuration instead.")
	}

	hostname = c.Hostname

	mta := mta.NewDefault(c,
		&Chain{handlers: []mta.Handler{}})
	go func() {
		<-sigc
		mta.Stop()
	}()
	err = mta.ListenAndServe()
	if err != nil {
		log.Errorln(err)
	}
}
